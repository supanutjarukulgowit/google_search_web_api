package service

import (
	"github.com/supanutjarukulgowit/google_search_web_api/database"
	"github.com/supanutjarukulgowit/google_search_web_api/model"
	"github.com/supanutjarukulgowit/google_search_web_api/util"
)

type UserService struct {
	PostgreSQLConnect model.PostgreSQLConnect
	PostgreSQL        *database.PostgreSQL
}

func NewUserService(postgreSQL interface{}) (*UserService, error) {
	var pConnect model.PostgreSQLConnect
	err := util.InterfaceToStruct(postgreSQL, &pConnect)
	if err != nil {
		return nil, err
	}

	return &UserService{
		PostgreSQLConnect: pConnect,
		PostgreSQL:        database.NewPostgreSQL(),
	}, nil
}
