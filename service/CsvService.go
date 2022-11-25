package service

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"strings"
	"time"

	"github.com/supanutjarukulgowit/google_search_web_api/database"
	"github.com/supanutjarukulgowit/google_search_web_api/model"
	"github.com/supanutjarukulgowit/google_search_web_api/repository"
	"github.com/supanutjarukulgowit/google_search_web_api/static"
	"github.com/supanutjarukulgowit/google_search_web_api/util"
)

type CsvService struct {
	Pg           *database.PostgreSQL
	PgConnection *model.PostgreSQLConnect
}

// var keywordsColumn = []string{"keyword_list"}

func NewCsvService(postgreSQL interface{}) (*CsvService, error) {
	var pConnect model.PostgreSQLConnect
	err := util.InterfaceToStruct(postgreSQL, &pConnect)
	if err != nil {
		return nil, err
	}
	return &CsvService{
		Pg:           database.NewPostgreSQL(),
		PgConnection: &pConnect,
	}, nil
}

func (h *CsvService) DownloadTemplate() (*model.DownloadTemplateResponse, error) {
	data, err := ioutil.ReadFile("../../templates/keyword_list.csv")
	if err != nil {
		return nil, fmt.Errorf("ConnectPostgreSQLGorm error : %s", err.Error())
	}
	return &model.DownloadTemplateResponse{
		Base64: base64.StdEncoding.EncodeToString(data),
	}, nil
}

func (h *CsvService) UploadFile(form *multipart.Form, gSearchApiKey string) (map[string]*model.GoogleSearchApiDetailDb, string, string, string, error) {
	db, err := h.Pg.ConnectPostgreSQLGorm(h.PgConnection.Host, h.PgConnection.User, h.PgConnection.Password, h.PgConnection.Database, h.PgConnection.Port)
	if err != nil {
		return nil, "", "", "", fmt.Errorf("ConnectPostgreSQLGorm error : %s", err.Error())
	}
	//validate
	if _, ok := form.File["files"]; !ok {
		return nil, "", "", static.INVALID_PARAMS, fmt.Errorf("file not found")
	}
	files := form.File["files"]
	if len(files) != 1 {
		//fix file = 1
		return nil, "", "", static.INVALID_PARAMS, fmt.Errorf("amount of file not correct")
	}
	if _, ok := form.Value["user_id"]; !ok {
		return nil, "", "", static.INVALID_PARAMS, fmt.Errorf("user_id is required")
	}
	userID := form.Value["user_id"]
	user := &model.User{}
	db.Where("id = ?", userID).First(&user)
	if user.Id == "" {
		return nil, "", "", static.USER_NOT_FOUND, fmt.Errorf("user_id not found")
	}
	//get keywords
	keywords, fileName, errCode, err := h.extractKeyWordsFromFile(files)
	if errCode != "" || err != nil {
		return nil, "", "", errCode, fmt.Errorf("extractKeyWordsFromFile error : %s", err.Error())
	}
	searchedKeywords, err := repository.GetSearchedKeyword(db, keywords)
	if err != nil {
		return nil, "", "", "", fmt.Errorf("getSearchedKeyword error : %s", err.Error())
	}
	//using only new keyword to prevent limitation
	//but still save data of found keywords
	searchID, _ := util.GetUUID() //id of each search
	foundKeywords, newKeywords, newKeywordList := h.filterKeywords(searchedKeywords, keywords, searchID, user.Id)
	err = repository.SaveSearchData(db, fileName, user.Id, searchID)
	if err != nil {
		return nil, "", "", "", fmt.Errorf("UploadFile|saveSearchData error : %s", err.Error())
	}
	if len(foundKeywords) != 0 {
		//save foundKeywords
		foundKData, err := repository.GetFoundKeywords(db, foundKeywords, user.Id, searchID)
		if err != nil {
			return nil, "", "", "", fmt.Errorf("UploadFile|getFoundKeywords error : %s", err.Error())
		}
		r := db.CreateInBatches(&foundKData, 50)
		if r.Error != nil {
			return nil, "", "", "", fmt.Errorf("UploadFile|save|foundKData| error : %s", err.Error())
		}
	}
	if len(newKeywordList) != 0 {
		r := db.CreateInBatches(&newKeywordList, 50)
		if r.Error != nil {
			return nil, "", "", "", fmt.Errorf("UploadFile|save|newKeywordList| error : %s", err.Error())
		}
	}
	return newKeywords, user.Id, searchID, "", nil
}

func (h *CsvService) extractKeyWordsFromFile(files []*multipart.FileHeader) ([]string, string, string, error) {
	keywords := make([]string, 0)
	//loop just incase for multiple files (for now file length 1)
	for _, file := range files {
		var records [][]string
		var errCode string
		var err error
		readAble := false

		if strings.Contains(file.Filename, "csv") {
			var keywordsColumn = []string{"keyword_list"}
			records, errCode, err = util.ReadUploadCsvFile(file, keywordsColumn)
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
				if record[0] != "" {
					keywords = append(keywords, record[0])
				}
			}
		}
	}
	//validate amount of keywords
	if len(keywords) > 100 {
		return nil, "", static.ERROR_AMOUNT_OF_SEARCH, fmt.Errorf("invalid amount of keywords")
	}
	return keywords, files[0].Filename, "", nil
}

func (h *CsvService) filterKeywords(searchedKeywords, keywords []string, searchID, userID string) (map[string]string, map[string]*model.GoogleSearchApiDetailDb, []model.GoogleSearchApiDetailDb) {
	foundKeyword := make(map[string]string, 0)
	newKeywordMap := make(map[string]*model.GoogleSearchApiDetailDb, 0)
	newKeywordList := make([]model.GoogleSearchApiDetailDb, 0)
	found := false
	for _, k := range keywords {
		found = false
		for _, sK := range searchedKeywords {
			if sK == k {
				found = true
				foundKeyword[sK] = sK
				break
			}
		}
		if !found {
			uuid, _ := util.GetUUID()
			nk := &model.GoogleSearchApiDetailDb{
				Id:          uuid,
				UserId:      userID,
				SearchId:    searchID,
				Status:      "pending",
				CreatedDate: time.Now(),
				Keyword:     k,
			}
			newKeywordMap[k] = nk
			newKeywordList = append(newKeywordList, *nk)
		}
	}
	return foundKeyword, newKeywordMap, newKeywordList
}
