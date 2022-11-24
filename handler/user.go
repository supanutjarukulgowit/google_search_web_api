package handler

import (
	"io/ioutil"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/supanutjarukulgowit/google_search_web_api/di"
	"github.com/supanutjarukulgowit/google_search_web_api/model"
	"github.com/supanutjarukulgowit/google_search_web_api/static"
	"github.com/supanutjarukulgowit/google_search_web_api/util"
)

type UserHandler interface {
	SignIn(c echo.Context) error
	SignUp(c echo.Context) error
}

type userHandler struct {
	BaseHandler
}

func NewUserHandler(postgreSQL interface{}) (UserHandler, error) {
	return &userHandler{}, nil
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
		errCode, err := userService.SignUp(&req)
		if errCode != "" {
			return util.GenError(c, static.USER_ALREADY_SIGN_UP, "", static.USER_ALREADY_SIGN_UP, http.StatusBadRequest)
		}
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
		token, errCode, err := userService.SignIn(&req)
		if errCode == "ERROR_004" {
			return util.GenError(c, errCode, static.ERROR_DESC["USER_NOT_FOUND"], errCode, http.StatusUnauthorized)
		} else if errCode == "ERROR_005" {
			return util.GenError(c, errCode, static.ERROR_DESC["USER_WRONG_PASSWORD"], errCode, http.StatusUnauthorized)
		}
		if err != nil {
			return util.GenError(c, static.USER_AUTHEN_ERROR, "SignUp error : "+err.Error(), static.USER_AUTHEN_ERROR, http.StatusInternalServerError)
		}
		return token
	}
	return h.RunProcess(c, convertFunc)
}
