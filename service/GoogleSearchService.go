package service

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/mitchellh/mapstructure"
	g "github.com/serpapi/google-search-results-golang"
	"github.com/supanutjarukulgowit/google_search_web_api/database"
	"github.com/supanutjarukulgowit/google_search_web_api/model"
	"github.com/supanutjarukulgowit/google_search_web_api/repository"
	"github.com/supanutjarukulgowit/google_search_web_api/util"
)

type GoogleSearchService struct {
	Pg            *database.PostgreSQL
	PgConnection  *model.PostgreSQLConnect
	GSearchApiKey string
}

func NewGoogleSearchService(postgreSQL interface{}) (*GoogleSearchService, error) {
	var pConnect model.PostgreSQLConnect
	err := util.InterfaceToStruct(postgreSQL, &pConnect)
	if err != nil {
		return nil, err
	}
	return &GoogleSearchService{
		Pg:           database.NewPostgreSQL(),
		PgConnection: &pConnect,
	}, nil
}

func (h *GoogleSearchService) GetGoogleSearchApi(keywords map[string]string, gSearchApiKey, userId, searchID string) error {
	db, err := h.Pg.ConnectPostgreSQLGorm(h.PgConnection.Host, h.PgConnection.User, h.PgConnection.Password, h.PgConnection.Database, h.PgConnection.Port)
	if err != nil {
		return fmt.Errorf("ConnectPostgreSQLGorm error : %s", err.Error())
	}
	poolSize := 3
	var wg sync.WaitGroup
	var mu sync.Mutex
	sResults := make([]model.GoogleSearchApiresponse, 0)
	errLog := make([]model.GoogleSearchErrorLog, 0)
	if len(keywords) != 0 {
		wg.Add(poolSize)
		ch := make(chan string, len(keywords))
		for thread := 1; thread <= poolSize; thread++ {
			go func(apiKey string) {
				defer wg.Done()
				for k := range ch {
					parameter := map[string]string{
						"q":       k,
						"engine":  "google",
						"api_key": gSearchApiKey,
					}
					search := g.NewGoogleSearch(parameter, gSearchApiKey)
					result, err := search.GetJSON()
					if err != nil {
						h.handleSaveLog(err, k, "search.GetJSON()", "", &errLog)
					}
					mu.Lock()
					var s model.GoogleSearchApiresponse
					mapstructure.Decode(result, &s)
					getHtmlCode, err := http.Get(s.SearchMetadata.GoogleUrl)
					if err != nil {
						s.SearchMetadata.HtmlCode = "cannot get html code on http get"
						h.handleSaveLog(err, k, "http.Get(s.SearchMetadata.GoogleUrl)", "", &errLog)
					}
					defer getHtmlCode.Body.Close()
					html, err := ioutil.ReadAll(getHtmlCode.Body)
					if err != nil {
						s.SearchMetadata.HtmlCode = "cannot get html code on ReadAll"
						h.handleSaveLog(err, k, "ioutil.ReadAll", "", &errLog)
					}
					s.SearchMetadata.HtmlCode = string(html)
					jsonStr, err := json.Marshal(result)
					if err != nil {
						h.handleSaveLog(err, k, "json.Marshal", "", &errLog)
					}
					s.TotalLinks = strings.Count(string(jsonStr), "https://")
					s.Keyword = k
					sResults = append(sResults, s)
					mu.Unlock()
				}
			}(gSearchApiKey)
		}

		for _, v := range keywords {
			ch <- v
		}
		close(ch)
		wg.Wait()
		if len(errLog) != 0 {
			db.CreateInBatches(errLog, 50)
		}
		err = repository.SaveSearchDataDetail(db, sResults, userId, searchID)
		if err != nil {
			return fmt.Errorf("UploadFile|saveSearchDataDetail error : %s", err.Error())
		}
	}
	return nil
}

func (h *GoogleSearchService) handleSaveLog(err error, keyword, action, userId string, errLog *[]model.GoogleSearchErrorLog) {
	id, _ := util.GetUUID()
	errObj := model.GoogleSearchErrorLog{
		Id:          id,
		ErrMessage:  fmt.Sprintf("search key word error %s for keywod : %s", err.Error(), keyword),
		Action:      action,
		CreatedDate: time.Now(),
	}
	*errLog = append(*errLog, errObj)
}
