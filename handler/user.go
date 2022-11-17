package handler

import (
	"fmt"

	"github.com/labstack/echo/v4"
	"github.com/supanutjarukulgowit/google_search_web_api/database"
	"github.com/supanutjarukulgowit/google_search_web_api/model"
	"github.com/supanutjarukulgowit/google_search_web_api/util"
)

type UserHandler interface {
	SignIn(c echo.Context) error
	SignUp(c echo.Context) error
}

type userHandler struct {
	PostgreSQLConnect model.PostgreSQLConnect
	PostgreSQL        *database.PostgreSQL
}

func NewUserHandler(postgreSQL interface{}) (UserHandler, error) {
	var pConnect model.PostgreSQLConnect
	err := util.InterfaceToStruct(postgreSQL, &pConnect)
	if err != nil {
		return nil, err
	}

	return &userHandler{
		PostgreSQLConnect: pConnect,
		PostgreSQL:        database.NewPostgreSQL(),
	}, nil

}

func (h *userHandler) SignIn(c echo.Context) error {

	db, err := h.PostgreSQL.ConnectPostgreSQL(h.PostgreSQLConnect.Host, h.PostgreSQLConnect.User, h.PostgreSQLConnect.Password, h.PostgreSQLConnect.Database, h.PostgreSQLConnect.Port)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(db)
	return nil
}

func (h *userHandler) SignUp(c echo.Context) error {

	db, err := h.PostgreSQL.ConnectPostgreSQL(h.PostgreSQLConnect.Host, h.PostgreSQLConnect.User, h.PostgreSQLConnect.Password, h.PostgreSQLConnect.Database, h.PostgreSQLConnect.Port)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(db)
	return nil
}
