package di

import (
	"github.com/supanutjarukulgowit/google_search_web_api/configuration"
	"github.com/supanutjarukulgowit/google_search_web_api/service"
)

var (
	_config        *configuration.Configuration
	userService    *service.UserService
	keywordService *service.KeywordService
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
