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
	"gorm.io/gorm"
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

type searchFunc func(parameter map[string]string, gSearchApiKey string) g.Search

func (h *GoogleSearchService) GetGoogleSearchApi(sFunc searchFunc, db *gorm.DB, keywordsMap map[string]*model.GoogleSearchApiDetailDb, gSearchApiKey, userId, searchID string) {
	poolSize := 3
	var wg sync.WaitGroup
	var mu sync.Mutex
	errLog := make([]model.GoogleSearchErrorLog, 0)
	if len(keywordsMap) != 0 {
		wg.Add(poolSize)
		ch := make(chan string, len(keywordsMap))
		for thread := 1; thread <= poolSize; thread++ {
			go func(apiKey string, keywordsMap map[string]*model.GoogleSearchApiDetailDb) {
				defer wg.Done()
				for k := range ch {
					parameter := map[string]string{
						"q":       k,
						"engine":  "google",
						"api_key": gSearchApiKey,
					}
					search := sFunc(parameter, gSearchApiKey)
					if _, ok := search.Parameter["test"]; !ok {
						result, err := search.GetJSON()
						if err != nil {
							h.handleSaveLog(err, k, "search.GetJSON()", "", &errLog, keywordsMap)
						}
						mu.Lock()
						var s model.GoogleSearchApiresponse
						mapstructure.Decode(result, &s)
						getHtmlCode, err := http.Get(s.SearchMetadata.GoogleUrl)
						if err != nil {
							h.handleSaveLog(err, k, "http.Get(s.SearchMetadata.GoogleUrl)", "", &errLog, keywordsMap)
						}
						defer getHtmlCode.Body.Close()
						html, err := ioutil.ReadAll(getHtmlCode.Body)
						if err != nil {
							h.handleSaveLog(err, k, "ioutil.ReadAll", "", &errLog, keywordsMap)
						}
						jsonStr, err := json.Marshal(result)
						if err != nil {
							h.handleSaveLog(err, k, "json.Marshal", "", &errLog, keywordsMap)
						}
						searchDetail := model.GoogleSearchApiDetailDb{
							Id:            keywordsMap[k].Id,
							SearchId:      keywordsMap[k].SearchId,
							UserId:        keywordsMap[k].UserId,
							CreatedDate:   keywordsMap[k].CreatedDate,
							Keyword:       k,
							AdWords:       len(s.Ads),
							Links:         strings.Count(string(jsonStr), "https://"),
							HTMLLink:      s.SearchMetadata.GoogleUrl,
							SearchResults: s.SearchInformation.TotalResults,
							TimeTaken:     s.SearchInformation.TimeTakenDisplayed,
							RawHTML:       html,
							Status:        "success",
						}
						keywordsMap[k] = &searchDetail
						mu.Unlock()
					} else {
						searchDetail := model.GoogleSearchApiDetailDb{
							Id:      "",
							Keyword: k,
							Status:  "success",
						}
						keywordsMap[k] = &searchDetail
					}
				}
			}(gSearchApiKey, keywordsMap)
		}

		for k, _ := range keywordsMap {
			ch <- k
		}
		close(ch)
		wg.Wait()
		if len(errLog) != 0 {
			db.CreateInBatches(errLog, 50)
		}
		err := repository.UpdateSearchDataDetail(keywordsMap, db)
		if err != nil {
			fmt.Println("UpdateSearchDataDetail error : %s user_id: %s searchID : %s", err.Error(), userId, searchID)
		}
	}
}

func (h *GoogleSearchService) handleSaveLog(err error, keyword, action, userId string, errLog *[]model.GoogleSearchErrorLog, keywordsMap map[string]*model.GoogleSearchApiDetailDb) {
	id, _ := util.GetUUID()
	errObj := model.GoogleSearchErrorLog{
		Id:          id,
		ErrMessage:  fmt.Sprintf("search key word error : %s for keywod : %s", err.Error(), keyword),
		Action:      action,
		CreatedDate: time.Now(),
	}
	errKeyword := &model.GoogleSearchApiDetailDb{
		Status: "failed",
		ErrMsg: fmt.Sprintf("search key word error : %s for keywod : %s", err.Error(), keyword),
	}
	keywordsMap[keyword] = errKeyword
	*errLog = append(*errLog, errObj)
}
