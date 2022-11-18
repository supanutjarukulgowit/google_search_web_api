package handler

import (
	"io/ioutil"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/supanutjarukulgowit/google_search_web_api/di"
	"github.com/supanutjarukulgowit/google_search_web_api/model"
	"github.com/supanutjarukulgowit/google_search_web_api/static"
	"github.com/supanutjarukulgowit/google_search_web_api/util"
)

type UserHandler interface {
	SignIn(c echo.Context) error
	SignUp(c echo.Context) error
	User(c echo.Context) error
	SignOut(c echo.Context) error
}

type userHandler struct {
	// PostgreSQLConnect model.PostgreSQLConnect
	// PostgreSQL        *database.PostgreSQL
	// DB *sql.DB
	// DBg *gorm.DB
	BaseHandler
}

func NewUserHandler(postgreSQL interface{}) (UserHandler, error) {
	var pConnect model.PostgreSQLConnect
	err := util.InterfaceToStruct(postgreSQL, &pConnect)
	if err != nil {
		return nil, err
	}
	// p := database.NewPostgreSQL()
	// db, err := p.ConnectPostgreSQLGorm(pConnect.Host, pConnect.User, pConnect.Password, pConnect.Database, pConnect.Port)
	// if err != nil {
	// 	return nil, err
	// }

	return &userHandler{
		// DBg: db,
	}, nil
}

func (h *userHandler) SignUp(c echo.Context) error {
	body, _ := ioutil.ReadAll(c.Request().Body)
	var req model.SignUpRequest
	err := util.ByteToStruct(body, &req)
	if err != nil {
		respErr := util.GenError(c, static.INVALID_PARAMS, "SignIn error: "+err.Error(), static.INVALID_PARAMS, http.StatusBadRequest)
		return c.JSON(http.StatusBadRequest, respErr)
	}
	err = util.ValidatorParam(req)
	if err != nil {
		respErr := util.GenError(c, static.INVALID_PARAMS, "SignIn error: "+err.Error(), static.INVALID_PARAMS, http.StatusBadRequest)
		return c.JSON(http.StatusBadRequest, respErr)
	}

	convertFunc := func() interface{} {
		userService, err := di.GetUserService()
		if err != nil {
			return util.GenError(c, static.INTERNAL_SERVER_ERROR, "GetUserService error : "+err.Error(), static.INTERNAL_SERVER_ERROR, http.StatusInternalServerError)
		}
		err = userService.SignUp(&req)
		if err != nil {
			return util.GenError(c, static.USER_AUTHEN_ERROR, "SignIn error : "+err.Error(), static.INTERNAL_SERVER_ERROR, http.StatusInternalServerError)
		}
		return nil
	}
	return h.RunProcess(c, convertFunc)
}

func (h *userHandler) SignIn(c echo.Context) error {
	body, _ := ioutil.ReadAll(c.Request().Body)
	var req model.SignInRequest
	err := util.ByteToStruct(body, &req)
	if err != nil {
		respErr := util.GenError(c, static.INVALID_PARAMS, "SignUp error: "+err.Error(), static.INVALID_PARAMS, http.StatusBadRequest)
		return c.JSON(http.StatusBadRequest, respErr)
	}
	err = util.ValidatorParam(req)
	if err != nil {
		respErr := util.GenError(c, static.INVALID_PARAMS, "SignUp error: "+err.Error(), static.INVALID_PARAMS, http.StatusBadRequest)
		return c.JSON(http.StatusBadRequest, respErr)
	}

	convertFunc := func() interface{} {
		userService, err := di.GetUserService()
		if err != nil {
			return util.GenError(c, static.INTERNAL_SERVER_ERROR, "GetUserService error : "+err.Error(), static.INTERNAL_SERVER_ERROR, http.StatusInternalServerError)
		}
		cookie, errCode, err := userService.SignIn(&req)
		if errCode == "ERROR_004" {
			return util.GenError(c, errCode, static.ERROR_DESC["USER_NOT_FOUND"], errCode, http.StatusUnauthorized)
		} else if errCode == "ERROR_005" {
			return util.GenError(c, errCode, static.ERROR_DESC["USER_WRONG_PASSWORD"], errCode, http.StatusUnauthorized)
		}
		if err != nil {
			return util.GenError(c, static.USER_AUTHEN_ERROR, "SignUp error : "+err.Error(), static.USER_AUTHEN_ERROR, http.StatusInternalServerError)
		}
		c.SetCookie(cookie)
		return nil
	}
	return h.RunProcess(c, convertFunc)
}

func (h *userHandler) User(c echo.Context) error {
	convertFunc := func() interface{} {
		userService, err := di.GetUserService()
		if err != nil {
			return util.GenError(c, static.INTERNAL_SERVER_ERROR, "GetUserService error : "+err.Error(), static.INTERNAL_SERVER_ERROR, http.StatusInternalServerError)
		}
		cookie, err := c.Cookie("jwt")
		if err != nil {
			return util.GenError(c, static.USER_AUTHEN_ERROR, "GetUserService error : "+err.Error(), static.USER_AUTHEN_ERROR, http.StatusUnauthorized)
		}
		response, errCode, err := userService.User(cookie)
		if errCode == "ERROR_003" {
			return util.GenError(c, static.USER_AUTHEN_ERROR, "", static.USER_AUTHEN_ERROR, http.StatusInternalServerError)
		}
		if err != nil {
			return util.GenError(c, static.USER_AUTHEN_ERROR, "SignUp error : "+err.Error(), static.USER_AUTHEN_ERROR, http.StatusInternalServerError)
		}
		return response
	}
	return h.RunProcess(c, convertFunc)
}

func (h *userHandler) SignOut(c echo.Context) error {
	convertFunc := func() interface{} {
		cookie := &http.Cookie{
			Name:     "jwt",
			Value:    "",
			Expires:  time.Now().Add(-time.Hour),
			HttpOnly: true,
		}
		c.SetCookie(cookie)
		return nil
	}
	return h.RunProcess(c, convertFunc)
}
