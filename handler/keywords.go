package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/supanutjarukulgowit/google_search_web_api/database"
	"github.com/supanutjarukulgowit/google_search_web_api/di"
	"github.com/supanutjarukulgowit/google_search_web_api/model"
	"github.com/supanutjarukulgowit/google_search_web_api/static"
	"github.com/supanutjarukulgowit/google_search_web_api/util"
)

type KeywordsHandler interface {
	DownloadTemplate(c echo.Context) error
	UploadFile(c echo.Context) error
	GetKeywordList(c echo.Context) error
}

type keywordHandler struct {
	BaseHandler
	Pg           *database.PostgreSQL
	PgConnection *model.PostgreSQLConnect
}

func NewKeywordsHandler(postgreSQL interface{}) (KeywordsHandler, error) {
	var pConnect model.PostgreSQLConnect
	err := util.InterfaceToStruct(postgreSQL, &pConnect)
	if err != nil {
		return nil, err
	}

	return &keywordHandler{
		Pg:           database.NewPostgreSQL(),
		PgConnection: &pConnect,
	}, nil
}

func (h *keywordHandler) DownloadTemplate(c echo.Context) error {
	convertFunc := func() interface{} {
		keywordervice, err := di.GetKeywordService()
		if err != nil {
			return util.GenError(c, static.INTERNAL_SERVER_ERROR, "DownloadTemplate error : "+err.Error(), static.INTERNAL_SERVER_ERROR, http.StatusInternalServerError)
		}
		response, err := keywordervice.DownloadTemplate()
		if err != nil {
			return util.GenError(c, static.DOWNLOAD_TEMPLATE_ERROR, "DownloadTemplate error : "+err.Error(), static.INTERNAL_SERVER_ERROR, http.StatusInternalServerError)
		}
		return response
	}
	return h.RunProcess(c, convertFunc)
}

func (h *keywordHandler) UploadFile(c echo.Context) error {
	form, err := c.MultipartForm()
	if err != nil {
		respErr := util.GenError(c, static.INVALID_PARAMS, "UploadFile error: "+err.Error(), static.INVALID_PARAMS, http.StatusBadRequest)
		return c.JSON(http.StatusBadRequest, respErr)
	}
	convertFunc := func() interface{} {
		keywordervice, err := di.GetKeywordService()
		if err != nil {
			return util.GenError(c, static.INTERNAL_SERVER_ERROR, "GetKeywordService error : "+err.Error(), static.INTERNAL_SERVER_ERROR, http.StatusInternalServerError)
		}

		var googleSearchConfig model.GoogleSearchConfig
		err = util.LoadDBConfig("google_search_api.config", &googleSearchConfig, h.Pg, h.PgConnection)
		if err != nil {
			return util.GenError(c, static.CANNOT_LOAD_CONFIG, "UploadFile error : "+err.Error(), static.CANNOT_LOAD_CONFIG, http.StatusInternalServerError)
		}

		errCode, err := keywordervice.UploadFile(form, &googleSearchConfig)
		if errCode != "" {
			if errCode == "ERR_001" {
				return util.GenError(c, errCode, "UploadFile error : "+err.Error(), errCode, http.StatusBadRequest)
			} else if errCode == "ERROR_004" {
				return util.GenError(c, errCode, "UploadFile error : "+err.Error(), errCode, http.StatusUnauthorized)
			} else {
				return util.GenError(c, errCode, "UploadFile error : "+err.Error(), errCode, http.StatusBadRequest)
			}
		}
		if err != nil {
			return util.GenError(c, static.UPLOAD_TEMPLATE_ERROR, "UploadFile error : "+err.Error(), static.INTERNAL_SERVER_ERROR, http.StatusInternalServerError)
		}
		return nil
	}
	return h.RunProcess(c, convertFunc)
}

func (h *keywordHandler) GetKeywordList(c echo.Context) error {
	userID := c.Request().Header.Get("user_id")
	if userID == "" {
		respErr := util.GenError(c, static.INVALID_PARAMS, "", static.INVALID_PARAMS, http.StatusUnauthorized)
		return c.JSON(http.StatusBadRequest, respErr)
	}
	convertFunc := func() interface{} {
		keywordervice, err := di.GetKeywordService()
		if err != nil {
			return util.GenError(c, static.INTERNAL_SERVER_ERROR, "GetKeywordService error : "+err.Error(), static.INTERNAL_SERVER_ERROR, http.StatusInternalServerError)
		}
		response, err := keywordervice.GetKeywordList(userID)
		if err != nil {
			return util.GenError(c, static.GET_KEYWORD_ERROR, "GetKeywordList error : "+err.Error(), static.INTERNAL_SERVER_ERROR, http.StatusInternalServerError)
		}
		return response
	}
	return h.RunProcess(c, convertFunc)
}
