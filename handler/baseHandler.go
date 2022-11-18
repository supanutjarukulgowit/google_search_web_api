package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/supanutjarukulgowit/google_search_web_api/common"
	"github.com/supanutjarukulgowit/google_search_web_api/util"
)

type BaseHandler struct {
}

type serviceFunc func() interface{}

func (h *BaseHandler) RunProcess(c echo.Context, sFunc serviceFunc) error {
	result := sFunc()
	return h.PostProcess(c, result)
}

func (h *BaseHandler) PostProcess(c echo.Context, response interface{}) error {
	var result common.TemplateResponse
	if v, ok := response.(common.TemplateResponse); ok {
		return c.JSON(v.HTTPStatus, v)
	} else {
		result = util.GenResponse(c, response)
	}
	return c.JSON(http.StatusOK, result)
}
