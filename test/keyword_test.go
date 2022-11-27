package test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/supanutjarukulgowit/google_search_web_api/configuration"
	"github.com/supanutjarukulgowit/google_search_web_api/di"
	"github.com/supanutjarukulgowit/google_search_web_api/model"
)

func TestGetKeywordList(t *testing.T) {
	config, err := configuration.LoadConfigFile("../config/config.local.json")
	if err != nil {
		t.Errorf("LoadConfigFile error: %s", err)
	} else {
		di.Init(config)
		keywordervice, err := di.GetKeywordService()
		if err != nil {
			t.Errorf("GetKeywordService error: %s", err)
		}
		db, mock, sqlDb := NewDatabase()
		defer sqlDb.Close()
		columns := []string{"id, keyword, search_id, ad_words, links, html_link, search_results, time_taken, cache, created_date, user_id, status, err_msg, test"}
		userIDMock := "6623CCE882D645338D1F5548F35B32FE"
		mock.ExpectQuery(regexp.QuoteMeta(`select id, keyword, ad_words, links, html_link, raw_html, search_results, time_taken, 
		created_date, cache from google_search_api_detail_dbs where user_id = ? ORDER BY created_date asc`)).WithArgs(userIDMock).WillReturnRows(mock.NewRows(columns))
		// now we execute our method
		resp, err := keywordervice.GetKeywordList(db, userIDMock)
		if err != nil {
			t.Errorf("error was not expected while updating stats: %s", err)
		}
		_ = resp
		// we make sure that all expectations were met
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
		assert.IsType(t, []model.GetKeywordListResponse{}, resp)
	}
}

func TestGetSearchKeyword(t *testing.T) {
	config, err := configuration.LoadConfigFile("../config/config.local.json")
	if err != nil {
		t.Errorf("LoadConfigFile error: %s", err)
	} else {
		di.Init(config)
		keywordervice, err := di.GetKeywordService()
		if err != nil {
			t.Errorf("GetKeywordService error: %s", err)
		}
		db, mock, sqlDb := NewDatabase()
		defer sqlDb.Close()
		columns := []string{"id, keyword, ad_words, links, html_link, raw_html, search_results, time_taken, created_date"}
		mockKeyword := &model.SearchKeywordRequest{
			Keyword: "TEST",
		}
		userIDMock := "6623CCE882D645338D1F5548F35B32FE"
		keywordEscaped := fmt.Sprintf("%%%s%%", mockKeyword.Keyword)
		mock.ExpectQuery(regexp.QuoteMeta(`select id, keyword, ad_words, links,
		html_link, raw_html, search_results, time_taken, created_date, cache from google_search_api_detail_dbs`)).WithArgs(keywordEscaped, userIDMock).WillReturnRows(mock.NewRows(columns))
		// now we execute our method
		resp, err := keywordervice.GetSearchKeyword(db, mockKeyword, userIDMock)
		if err != nil {
			t.Errorf("error was not expected while updating stats: %s", err)
		}
		_ = resp
		// we make sure that all expectations were met
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
		assert.IsType(t, []model.GetKeywordListResponse{}, resp)
	}
}
