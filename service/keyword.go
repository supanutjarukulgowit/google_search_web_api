package service

import (
	"database/sql"
	"encoding/base64"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/mitchellh/mapstructure"
	g "github.com/serpapi/google-search-results-golang"
	"github.com/supanutjarukulgowit/google_search_web_api/database"
	"github.com/supanutjarukulgowit/google_search_web_api/model"
	"github.com/supanutjarukulgowit/google_search_web_api/static"
	"github.com/supanutjarukulgowit/google_search_web_api/util"
	"gorm.io/gorm"
)

type KeywordService struct {
	Pg           *database.PostgreSQL
	PgConnection *model.PostgreSQLConnect
}

var keywordsColumn = []string{"keyword_list"}

func NewKeywordService(postgreSQL interface{}) (*KeywordService, error) {
	var pConnect model.PostgreSQLConnect
	err := util.InterfaceToStruct(postgreSQL, &pConnect)
	if err != nil {
		return nil, err
	}
	return &KeywordService{
		Pg:           database.NewPostgreSQL(),
		PgConnection: &pConnect,
	}, nil
}

func (h *KeywordService) DownloadTemplate() (*model.DownloadTemplateResponse, error) {
	data, err := ioutil.ReadFile("../../templates/keyword_list.csv")
	if err != nil {
		return nil, fmt.Errorf("ConnectPostgreSQLGorm error : %s", err.Error())
	}
	return &model.DownloadTemplateResponse{
		Base64: base64.StdEncoding.EncodeToString(data),
	}, nil
}

func (h *KeywordService) UploadFile(form *multipart.Form, googleSearchConfig *model.GoogleSearchConfig) (string, error) {
	db, err := h.Pg.ConnectPostgreSQLGorm(h.PgConnection.Host, h.PgConnection.User, h.PgConnection.Password, h.PgConnection.Database, h.PgConnection.Port)
	if err != nil {
		return "", fmt.Errorf("ConnectPostgreSQLGorm error : %s", err.Error())
	}
	//validate
	if _, ok := form.File["files"]; !ok {
		return static.INVALID_PARAMS, fmt.Errorf("file not found")
	}
	files := form.File["files"]
	if len(files) != 1 {
		//fix file = 1
		return static.INVALID_PARAMS, fmt.Errorf("amount of file not correct")
	}
	if _, ok := form.Value["user_id"]; !ok {
		return static.INVALID_PARAMS, fmt.Errorf("user_id is required")
	}
	userID := form.Value["user_id"]
	user := &model.User{}
	db.Where("id = ?", userID).First(&user)
	if user.Id == "" {
		return static.USER_NOT_FOUND, fmt.Errorf("user_id not found")
	}

	//get keywords
	keywords, fileName, errCode, err := h.extractKeyWordsFromFile(files)
	if errCode != "" || err != nil {
		return errCode, fmt.Errorf("extractKeyWordsFromFile error : %s", err.Error())
	}
	//get data from search api
	result, err := h.getGoogleSearchApi(keywords, googleSearchConfig)
	if err != nil {
		return "", fmt.Errorf("search.GetJSON() error : %s", err.Error())
	}
	//save
	err = h.saveSearchData(db, result, fileName, user.Id)
	if err != nil {
		return "", fmt.Errorf("UploadFile|saveSearchData error : %s", err.Error())
	}
	return "", nil
}

func (h *KeywordService) extractKeyWordsFromFile(files []*multipart.FileHeader) ([]string, string, string, error) {
	keywords := make([]string, 0)
	//loop just incase for multiple files (for now file length 1)
	for _, file := range files {
		var records [][]string
		var errCode string
		var err error
		readAble := false

		if strings.Contains(file.Filename, "csv") {
			records, errCode, err = h.readUploadFile(file)
			if err != nil {
				return nil, "", errCode, fmt.Errorf("readUploadFile error : %s", err.Error())
			}
			readAble = true
		} else {
			//using only csv
			return nil, "", static.INVALID_PARAMS, fmt.Errorf("invalid type of file")
		}
		if readAble {
			for _, record := range records {
				keywords = append(keywords, record[0])
			}
		}
	}
	//validate amount of keywords
	if len(keywords) > 100 {
		return nil, "", static.ERROR_AMOUNT_OF_SEARCH, fmt.Errorf("invalid amount of keywords")
	}
	return keywords, files[0].Filename, "", nil
}

func (h *KeywordService) readUploadFile(file *multipart.FileHeader) ([][]string, string, error) {
	src, err := file.Open()
	if err != nil {
		return nil, static.CANNOT_READ_FILE_ERROR, fmt.Errorf("open file error %s", err.Error())
	}
	defer src.Close()

	csvReader := csv.NewReader(src)
	records, err := csvReader.ReadAll()
	if err != nil {
		return nil, static.CANNOT_READ_FILE_ERROR, fmt.Errorf("read file error %s", err.Error())
	}
	if len(records) == 1 {
		return nil, static.FILE_NO_DATA_ERROR, fmt.Errorf("no data found")
	}

	if len(records[0]) < len(keywordsColumn) {
		return nil, static.COLUMN_COUNT_INVALID, fmt.Errorf("column count invalid")
	}

	for index, col := range records[0][0:len(keywordsColumn)] {
		if col != keywordsColumn[index] {
			return nil, static.COLUMN_NAME_INVALID, fmt.Errorf("column name invalid")
		}
	}
	return records[1:], "", nil
}

func (h *KeywordService) getGoogleSearchApi(keywords []string, googleSearchConfig *model.GoogleSearchConfig) ([]model.GoogleSearchApiresponse, error) {
	poolSize := 3
	var wg sync.WaitGroup
	var mu sync.Mutex
	wg.Add(poolSize)
	ch := make(chan string, len(keywords))
	// searchResults := make([]g.SearchResult, 0)
	sResults := make([]model.GoogleSearchApiresponse, 0)
	for thread := 1; thread <= poolSize; thread++ {
		go func(apiKey string) {
			defer wg.Done()
			for k := range ch {
				parameter := map[string]string{
					"q":       k,
					"engine":  "google",
					"api_key": apiKey,
				}
				search := g.NewGoogleSearch(parameter, googleSearchConfig.Apikey)
				result, err := search.GetJSON()
				if err != nil {
					fmt.Println(err)
				}
				mu.Lock()
				var s model.GoogleSearchApiresponse
				mapstructure.Decode(result, &s)
				getHtmlCode, err := http.Get(s.SearchMetadata.GoogleUrl)
				if err != nil {
					s.SearchMetadata.HtmlCode = "cannot get html code on http get"
				}
				defer getHtmlCode.Body.Close()
				html, err := ioutil.ReadAll(getHtmlCode.Body)
				if err != nil {
					s.SearchMetadata.HtmlCode = "cannot get html code on ReadAll"
				}
				s.SearchMetadata.HtmlCode = string(html)
				jsonStr, err := json.Marshal(result)
				if err != nil {
					fmt.Println(err)
				}
				s.TotalLinks = strings.Count(string(jsonStr), "https://")
				s.Keyword = k
				sResults = append(sResults, s)
				mu.Unlock()
			}
		}(googleSearchConfig.Apikey)
	}

	for _, k := range keywords {
		ch <- k
	}
	close(ch)
	wg.Wait()

	return sResults, nil
}

func (h *KeywordService) saveSearchData(db *gorm.DB, result []model.GoogleSearchApiresponse, fileName, userID string) error {
	id, _ := util.GetUUID()
	searchID, _ := util.GetUUID()
	search := model.GoogleSearchApiDb{
		Id:          id,
		SearchId:    searchID,
		UserId:      userID,
		FileName:    fileName + "_" + time.Now().Format("2006-01-02_15:04:05"),
		CreatedDate: time.Now(),
	}
	db.Create(&search)
	searchDetails := make([]model.GoogleSearchApiDetailDb, 0)
	for _, r := range result {
		detailID, _ := util.GetUUID()
		detail := model.GoogleSearchApiDetailDb{
			Id:            detailID,
			SearchId:      searchID,
			Keyword:       r.Keyword,
			AdWords:       len(r.Ads),
			Links:         r.TotalLinks,
			HTMLLink:      r.SearchMetadata.GoogleUrl,
			SearchResults: r.SearchInformation.TotalResults,
			TimeTaken:     r.SearchInformation.TimeTakenDisplayed,
			CreatedDate:   time.Now(),
			UserId:        userID,
			RawHTML:       []byte(r.SearchMetadata.HtmlCode),
		}
		searchDetails = append(searchDetails, detail)
	}
	db.CreateInBatches(&searchDetails, 50)
	return nil
}

func (h *KeywordService) GetKeywordList(userID string) ([]model.GetKeywordListResponse, error) {
	db, err := h.Pg.ConnectPostgreSQLGorm(h.PgConnection.Host, h.PgConnection.User, h.PgConnection.Password, h.PgConnection.Database, h.PgConnection.Port)
	if err != nil {
		return nil, fmt.Errorf("ConnectPostgreSQLGorm error : %s", err.Error())
	}
	details := []model.GetKeywordListResponse{}
	// Raw SQL

	rows, err := db.Raw(`select keyword, ad_words, links,
	html_link, raw_html, search_results, time_taken, created_date, cache from google_search_api_detail_dbs
	where user_id = ?`, userID).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var keyword sql.NullString
		var adWords sql.NullInt32
		var links sql.NullInt32
		var htmlLink sql.NullString
		var rawHtml sql.NullString
		var searchResults sql.NullInt32
		var timeTaken sql.NullFloat64
		var createdDate sql.NullTime
		var cache sql.NullString

		err := rows.Scan(&keyword, &adWords, &links, &htmlLink, &rawHtml, &searchResults, &timeTaken, &createdDate, &cache)
		if err != nil {
			fmt.Println(err)
			return nil, err
		}
		detail := model.GetKeywordListResponse{
			Keyword:       util.GetStringFromSQL(keyword),
			AdWords:       util.GetIntFromSQL(adWords),
			Links:         util.GetIntFromSQL(links),
			HTMLLink:      util.GetStringFromSQL(htmlLink),
			SearchResults: util.GetIntFromSQL(adWords),
			Cache:         util.GetStringFromSQL(cache),
			TimeTaken:     util.GetFloatFromSQL(timeTaken),
			RawHTML:       util.GetStringFromSQL(rawHtml),
		}
		if date := util.GetTimeFromSQL(createdDate); date != nil && !date.IsZero() {
			detail.CreatedDate = util.TimestampToString("", date.Unix())
		}
		details = append(details, detail)
	}
	return details, nil
}
