package di

import (
	"github.com/supanutjarukulgowit/google_search_web_api/configuration"
	"github.com/supanutjarukulgowit/google_search_web_api/service"
)

var (
	_config             *configuration.Configuration
	userService         *service.UserService
	keywordService      *service.KeywordService
	csvService          *service.CsvService
	googleSearchService *service.GoogleSearchService
)

//Init service
func Init(config *configuration.Configuration) {
	_config = config
}

func GetUserService() (*service.UserService, error) {
	if userService == nil {
		var err error
		userService, err = service.NewUserService(_config.PostgreSQL)
		return userService, err
	}

	return userService, nil
}

func GetKeywordService() (*service.KeywordService, error) {
	if keywordService == nil {
		var err error
		keywordService, err = service.NewKeywordService(_config.PostgreSQL)
		return keywordService, err
	}

	return keywordService, nil
}

func GetCsvService() (*service.CsvService, error) {
	if csvService == nil {
		var err error
		csvService, err = service.NewCsvService(_config.PostgreSQL)
		return csvService, err
	}

	return csvService, nil
}

func GetGoogleSearchService() (*service.GoogleSearchService, error) {
	if googleSearchService == nil {
		var err error
		googleSearchService, err = service.NewGoogleSearchService(_config.PostgreSQL)
		return googleSearchService, err
	}

	return googleSearchService, nil
}
