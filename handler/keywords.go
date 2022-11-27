package handler

import (
	"io/ioutil"
	"net/http"

	"github.com/labstack/echo/v4"
	g "github.com/serpapi/google-search-results-golang"
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
	GetSearchKeyword(c echo.Context) error
}

type keywordHandler struct {
	BaseHandler
	Pg            *database.PostgreSQL
	PgConnection  *model.PostgreSQLConnect
	GSearchApiKey string
}

type searchFunc func() interface{}

func NewKeywordsHandler(postgreSQL interface{}, gSearchApiKey string) (KeywordsHandler, error) {
	var pConnect model.PostgreSQLConnect
	err := util.InterfaceToStruct(postgreSQL, &pConnect)
	if err != nil {
		return nil, err
	}

	return &keywordHandler{
		Pg:            database.NewPostgreSQL(),
		PgConnection:  &pConnect,
		GSearchApiKey: gSearchApiKey,
	}, nil
}

func (h *keywordHandler) DownloadTemplate(c echo.Context) error {
	convertFunc := func() interface{} {
		csvService, err := di.GetCsvService()
		if err != nil {
			return util.GenError(c, static.INTERNAL_SERVER_ERROR, "GetCsvService error : "+err.Error(), static.INTERNAL_SERVER_ERROR, http.StatusInternalServerError)
		}
		response, err := csvService.DownloadTemplate()
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
		csvService, err := di.GetCsvService()
		if err != nil {
			return util.GenError(c, static.INTERNAL_SERVER_ERROR, "GetCsvService error : "+err.Error(), static.INTERNAL_SERVER_ERROR, http.StatusInternalServerError)
		}
		googleSearchService, err := di.GetGoogleSearchService()
		if err != nil {
			return util.GenError(c, static.INTERNAL_SERVER_ERROR, "GetGoogleSearchService error : "+err.Error(), static.INTERNAL_SERVER_ERROR, http.StatusInternalServerError)
		}
		db, err := h.Pg.ConnectPostgreSQLGorm(h.PgConnection.Host, h.PgConnection.User, h.PgConnection.Password, h.PgConnection.Database, h.PgConnection.Port)
		if err != nil {
			return util.GenError(c, static.INTERNAL_SERVER_ERROR, "GetGoogleSearchService error : "+err.Error(), static.INTERNAL_SERVER_ERROR, http.StatusInternalServerError)
		}

		newkeywords, userID, searchID, errCode, err := csvService.UploadFile(db, form)
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
		//upload new keyword and then update status by async processing
		searchFunc := func(parameter map[string]string, gSearchApiKey string) g.Search {
			search := g.NewGoogleSearch(parameter, gSearchApiKey)
			return search
		}
		go func() {
			googleSearchService.GetGoogleSearchApi(searchFunc, db, newkeywords, h.GSearchApiKey, userID, searchID)
		}()
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
		db, err := h.Pg.ConnectPostgreSQLGorm(h.PgConnection.Host, h.PgConnection.User, h.PgConnection.Password, h.PgConnection.Database, h.PgConnection.Port)
		if err != nil {
			return util.GenError(c, static.INTERNAL_SERVER_ERROR, "ConnectPostgreSQLGorm error : "+err.Error(), static.INTERNAL_SERVER_ERROR, http.StatusInternalServerError)
		}
		response, err := keywordervice.GetKeywordList(db, userID)
		if err != nil {
			return util.GenError(c, static.GET_KEYWORD_ERROR, "GetKeywordList error : "+err.Error(), static.INTERNAL_SERVER_ERROR, http.StatusInternalServerError)
		}
		return response
	}
	return h.RunProcess(c, convertFunc)
}

func (h *keywordHandler) GetSearchKeyword(c echo.Context) error {
	body, _ := ioutil.ReadAll(c.Request().Body)
	var req model.SearchKeywordRequest
	err := util.ByteToStruct(body, &req)
	if err != nil {
		respErr := util.GenError(c, static.INVALID_PARAMS, "GetSearchKeyword error: "+err.Error(), static.INVALID_PARAMS, http.StatusBadRequest)
		return c.JSON(http.StatusBadRequest, respErr)
	}
	err = util.ValidatorParam(req)
	if err != nil {
		respErr := util.GenError(c, static.INVALID_PARAMS, "GetSearchKeyword error: "+err.Error(), static.INVALID_PARAMS, http.StatusBadRequest)
		return c.JSON(http.StatusBadRequest, respErr)
	}
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
		db, err := h.Pg.ConnectPostgreSQLGorm(h.PgConnection.Host, h.PgConnection.User, h.PgConnection.Password, h.PgConnection.Database, h.PgConnection.Port)
		if err != nil {
			return util.GenError(c, static.INTERNAL_SERVER_ERROR, "ConnectPostgreSQLGorm error : "+err.Error(), static.INTERNAL_SERVER_ERROR, http.StatusInternalServerError)
		}
		response, err := keywordervice.GetSearchKeyword(db, &req, userID)
		if err != nil {
			return util.GenError(c, static.INTERNAL_SERVER_ERROR, "GetSearchKeyword error : "+err.Error(), static.INTERNAL_SERVER_ERROR, http.StatusInternalServerError)
		}
		return response
	}
	return h.RunProcess(c, convertFunc)
}
