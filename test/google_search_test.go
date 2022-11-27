package test

import (
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	g "github.com/serpapi/google-search-results-golang"
	"github.com/supanutjarukulgowit/google_search_web_api/configuration"
	"github.com/supanutjarukulgowit/google_search_web_api/di"
	"github.com/supanutjarukulgowit/google_search_web_api/model"
)

func TestGetGoogleSearchApi(t *testing.T) {
	config, err := configuration.LoadConfigFile("../config/config.local.json")
	if err != nil {
		t.Errorf("LoadConfigFile error: %s", err)
	} else {
		di.Init(config)
		googleSearchService, err := di.GetGoogleSearchService()
		if err != nil {
			t.Errorf("googleSearchService error: %s", err)
		}
		db, mock, sqlDb := NewDatabase()
		defer sqlDb.Close()
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta("UPDATE `google_search_api_detail_dbs`")).WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()
		searchFuncMock := func(parameter map[string]string, gSearchApiKey string) g.Search {
			parameter["test"] = "test"
			search := g.Search{
				Parameter: parameter,
			}
			return search
		}
		keywordMock := make(map[string]*model.GoogleSearchApiDetailDb)
		keywordMock["test"] = nil
		googleSearchService.GetGoogleSearchApi(searchFuncMock, db, keywordMock, "", "", "")
		if err != nil {
			t.Errorf("error was not expected while updating stats: %s", err)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	}
}
