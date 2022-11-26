package handler

import (
	"database/sql"
	"fmt"
	"log"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/supanutjarukulgowit/google_search_web_api/configuration"
	"github.com/supanutjarukulgowit/google_search_web_api/di"
	"github.com/tj/assert"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type mockGormDb struct {
}

func (m *mockGormDb) Raw(sql string, values ...interface{}) (tx *gorm.DB) {
	return
}

func (m *mockGormDb) Rows() (*sql.Rows, error) {
	r := &sql.Rows{}
	return r, nil
}

func TestGetKeywordList(t *testing.T) {
	config, err := configuration.LoadConfigFile("../config/config.local.json")
	if err != nil {
		assert.Nil(t, err)
	} else {
		di.Init(config)
		// h, _ := handler.NewKeywordsHandler(config, config.GoogleSearchApiKey)
		// keywordervice, err := di.GetKeywordService()
		// if err != nil {

		// }
		// mockGormDb := &mockGormDb{}
		// response, err := keywordervice.GetKeywordList(mockGormDb, "")
		// if err != nil {
		// 	assert.Nil(t, err)
		// }
		// fmt.Println(response)
	}
}

func TestGetKeywordListWithLib(t *testing.T) {
	config, err := configuration.LoadConfigFile("../config/config.local.json")
	if err != nil {
		assert.Nil(t, err)
	} else {
		di.Init(config)
		keywordervice, err := di.GetKeywordService()
		if err != nil {
			assert.Nil(t, err)
		}
		db, mock, sqlDb := NewDatabase()
		defer sqlDb.Close()
		// a SELECT VERSION() query will be run when gorm opens the database
		// so we need to expect that here
		columns := []string{"id, keyword, search_id, ad_words, links, html_link, search_results, time_taken, cache, created_date, user_id, status, err_msg"}
		userIDMock := "6623CCE882D645338D1F5548F35B32FE"
		mock.ExpectQuery(`select id, keyword, ad_words, links, html_link, raw_html, search_results, time_taken, created_date, cache from google_search_api_detail_dbs where user_id = ? ORDER BY created_date asc`).WithArgs(userIDMock).WillReturnRows(mock.NewRows(columns))

		// now we execute our method
		resp, err := keywordervice.GetKeywordList(db, userIDMock)
		if err != nil {
			t.Errorf("error was not expected while updating stats: %s", err)
		}
		fmt.Println(resp)

		// we make sure that all expectations were met
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	}

}

func NewDatabase() (*gorm.DB, sqlmock.Sqlmock, *sql.DB) {
	// get db and mock
	sqlDB, mock, err := sqlmock.New(
		sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp),
	)
	if err != nil {
		log.Fatalf("[sqlmock new] %s", err)
	}
	// defer sqlDB.Close()

	// create dialector
	dialector := mysql.New(mysql.Config{
		Conn:       sqlDB,
		DriverName: "mysql",
	})

	columns := []string{"version"}
	mock.ExpectQuery("SELECT VERSION()").WithArgs().WillReturnRows(
		mock.NewRows(columns).FromCSVString("1"),
	)
	// open the database
	db, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		log.Fatalf("[gorm open] %s", err)
	}

	return db, mock, sqlDB
}
