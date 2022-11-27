package test

import (
	"bytes"
	"database/sql/driver"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/supanutjarukulgowit/google_search_web_api/configuration"
	"github.com/supanutjarukulgowit/google_search_web_api/di"
	"github.com/supanutjarukulgowit/google_search_web_api/model"
)

type AnyString struct{}
type AnyTime struct{}

// Match satisfies sqlmock.Argument interface
func (a AnyString) Match(v driver.Value) bool {
	_, ok := v.(string)
	return ok
}

// Match satisfies sqlmock.Argument interface
func (a AnyTime) Match(v driver.Value) bool {
	_, ok := v.(time.Time)
	return ok
}

type Test struct {
	AnyString AnyString
	AnyTime   AnyTime
}

func TestUploadFile(t *testing.T) {
	config, err := configuration.LoadConfigFile("../config/config.local.json")
	if err != nil {
		t.Errorf("LoadConfigFile error: %s", err)
	} else {
		di.Init(config)
		csvService, err := di.GetCsvService()
		if err != nil {
			t.Errorf("GetCsvService error: %s", err)
		}
		path := "../templates/keyword_list_test_validate.csv"
		userIDMock := "6623CCE882D645338D1F5548F35B32FE"
		form, err := createMockFile(path, userIDMock)
		if err != nil {
			t.Errorf("createMockFile error: %s", err)
		}
		db, mock, sqlDb := NewDatabase()
		defer sqlDb.Close()
		userCol := []string{"id, username, password, first_name, last_name, created_date"}
		keywordCol := []string{"keyword"}

		mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `users`")).WithArgs(userIDMock).WillReturnRows(mock.NewRows(userCol))
		mock.ExpectQuery(regexp.QuoteMeta("select distinct keyword from google_search_api_detail_dbs")).WillReturnRows(mock.NewRows(keywordCol))
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `google_search_api_dbs`")).WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `google_search_api_detail_dbs`")).WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		// now we execute our method
		resp, userID, searchID, errCode, err := csvService.UploadFile(db, form)
		if err != nil {
			t.Errorf("error was not expected while UploadFile: %s", err)
		}
		// we make sure that all expectations were met
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
		assert.IsType(t, map[string]*model.GoogleSearchApiDetailDb{}, resp)
		assert.IsType(t, "", userID)
		assert.IsType(t, searchID, searchID)
		if errCode != "" {
			t.Errorf("error was not expected errCode: %s", errCode)
		}
	}
}

func TestUploadFileNotFound(t *testing.T) {
	config, err := configuration.LoadConfigFile("../config/config.local.json")
	if err != nil {
		t.Errorf("LoadConfigFile error: %s", err)
	} else {
		di.Init(config)
		csvService, err := di.GetCsvService()
		if err != nil {
			t.Errorf("GetCsvService error: %s", err)
		}
		path := "../templates/keyword_list_test_validate.csv"
		userIDMock := "6623CCE882D645338D1F5548F35B32FE"
		form, err := createMockFileNotFound(path, userIDMock)
		if err != nil {
			t.Errorf("createMockFile error: %s", err)
		}
		resp, _, _, errCode, err := csvService.UploadFile(nil, form)
		if err != nil {
			assert.IsType(t, map[string]*model.GoogleSearchApiDetailDb{}, resp)
			assert.Equal(t, "ERR_001", errCode)
		} else {
			t.Errorf("UploadFile should throw an error invalid param")
		}

	}
}

func TestUploadFileNotCSV(t *testing.T) {
	config, err := configuration.LoadConfigFile("../config/config.local.json")
	if err != nil {
		t.Errorf("LoadConfigFile error: %s", err)
	} else {
		di.Init(config)
		csvService, err := di.GetCsvService()
		if err != nil {
			t.Errorf("GetCsvService error: %s", err)
		}
		path := "../templates/keyword_list_test_validate.txt"
		userIDMock := "6623CCE882D645338D1F5548F35B32FE"
		form, err := createMockFile(path, userIDMock)
		if err != nil {
			t.Errorf("createMockFile error: %s", err)
		}
		db, mock, sqlDb := NewDatabase()
		defer sqlDb.Close()
		userCol := []string{"id, username, password, first_name, last_name, created_date"}

		mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `users`")).WithArgs(userIDMock).WillReturnRows(mock.NewRows(userCol))

		resp, _, _, errCode, err := csvService.UploadFile(db, form)
		if err != nil {
			assert.IsType(t, map[string]*model.GoogleSearchApiDetailDb{}, resp)
			assert.Equal(t, "ERR_001", errCode)
		} else {
			t.Errorf("UploadFile should throw an error invalid param")
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	}
}

func createMockFileNotFound(path, userIDMock string) (*multipart.Form, error) {
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("testNotFound", path) //proper name is files
	if err != nil {
		return nil, fmt.Errorf("CreateFormFile error : %s", err.Error())
	}
	sample, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("os.Open error : %s", err.Error())
	}
	_, err = io.Copy(part, sample)
	if err != nil {
		return nil, fmt.Errorf("io.Copy error : %s", err.Error())
	}
	writer.Close()
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", body)
	req.Header.Set(echo.HeaderContentType, writer.FormDataContentType())
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	form, err := c.MultipartForm()
	if err != nil {
		return nil, fmt.Errorf("MultipartForm error : %s", err.Error())
	}
	form.Value["user_id"] = []string{userIDMock}
	return form, nil
}

func createMockFile(path, userIDMock string) (*multipart.Form, error) {
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("files", path)
	if err != nil {
		return nil, fmt.Errorf("CreateFormFile error : %s", err.Error())
	}
	sample, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("os.Open error : %s", err.Error())
	}
	_, err = io.Copy(part, sample)
	if err != nil {
		return nil, fmt.Errorf("io.Copy error : %s", err.Error())
	}
	writer.Close()
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", body)
	req.Header.Set(echo.HeaderContentType, writer.FormDataContentType())
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	form, err := c.MultipartForm()
	if err != nil {
		return nil, fmt.Errorf("MultipartForm error : %s", err.Error())
	}
	form.Value["user_id"] = []string{userIDMock}
	return form, nil
}
