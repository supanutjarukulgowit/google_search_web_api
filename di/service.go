package di

import (
	"github.com/supanutjarukulgowit/google_search_web_api/configuration"
	"github.com/supanutjarukulgowit/google_search_web_api/service"
)

var (
	_config     *configuration.Configuration
	userService *service.UserService
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
