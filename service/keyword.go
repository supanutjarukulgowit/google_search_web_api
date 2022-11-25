package service

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/supanutjarukulgowit/google_search_web_api/database"
	"github.com/supanutjarukulgowit/google_search_web_api/model"
	"github.com/supanutjarukulgowit/google_search_web_api/util"
)

type KeywordService struct {
	Pg           *database.PostgreSQL
	PgConnection *model.PostgreSQLConnect
}

func NewKeywordService(postgreSQL interface{}) (*KeywordService, error) {
	var pConnect model.PostgreSQLConnect
	err := util.InterfaceToStruct(postgreSQL, &pConnect)
	if err != nil {
		return nil, err
	}
	return &KeywordService{
		Pg:           database.NewPostgreSQL(),
		PgConnection: &pConnect,
	}, nil
}

func (h *KeywordService) GetKeywordList(userID string) ([]model.GetKeywordListResponse, error) {
	db, err := h.Pg.ConnectPostgreSQLGorm(h.PgConnection.Host, h.PgConnection.User, h.PgConnection.Password, h.PgConnection.Database, h.PgConnection.Port)
	if err != nil {
		return nil, fmt.Errorf("ConnectPostgreSQLGorm error : %s", err.Error())
	}
	details := []model.GetKeywordListResponse{}
	// Raw SQL

	rows, err := db.Raw(`select id, keyword, ad_words, links,
	html_link, raw_html, search_results, time_taken, created_date, cache from google_search_api_detail_dbs
	where user_id = ? ORDER BY created_date asc `, userID).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var id sql.NullString
		var keyword sql.NullString
		var adWords sql.NullInt32
		var links sql.NullInt32
		var htmlLink sql.NullString
		var rawHtml sql.NullString
		var searchResults sql.NullInt64
		var timeTaken sql.NullFloat64
		var createdDate sql.NullTime
		var cache sql.NullString

		err := rows.Scan(&id, &keyword, &adWords, &links, &htmlLink, &rawHtml, &searchResults, &timeTaken, &createdDate, &cache)
		if err != nil {
			return nil, err
		}
		detail := model.GetKeywordListResponse{
			Id:            util.GetStringFromSQL(id),
			Keyword:       util.GetStringFromSQL(keyword),
			AdWords:       util.GetIntFromSQL(adWords),
			Links:         util.GetIntFromSQL(links),
			HTMLLink:      util.GetStringFromSQL(htmlLink),
			SearchResults: util.GetInt64FromSQL(searchResults),
			Cache:         util.GetStringFromSQL(cache),
			TimeTaken:     util.GetFloatFromSQL(timeTaken),
			RawHTML:       util.GetStringFromSQL(rawHtml),
		}
		if date := util.GetTimeFromSQL(createdDate); date != nil && !date.IsZero() {
			detail.CreatedDate = util.TimestampToString("", date.Unix())
		}
		details = append(details, detail)
	}
	return details, nil
}

func (h *KeywordService) GetSearchKeyword(req *model.SearchKeywordRequest, userID string) ([]model.GetKeywordListResponse, error) {
	db, err := h.Pg.ConnectPostgreSQLGorm(h.PgConnection.Host, h.PgConnection.User, h.PgConnection.Password, h.PgConnection.Database, h.PgConnection.Port)
	if err != nil {
		return nil, fmt.Errorf("ConnectPostgreSQLGorm error : %s", err.Error())
	}
	details := []model.GetKeywordListResponse{}
	// Raw SQL
	k := fmt.Sprintf("%%%s%%", strings.ToUpper(req.Keyword))
	rows, err := db.Raw(`select id, keyword, ad_words, links,
	html_link, raw_html, search_results, time_taken, created_date, cache from google_search_api_detail_dbs
	where UPPER(keyword) like ? and user_id = ? ORDER BY created_date asc `, k, userID).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var id sql.NullString
		var keyword sql.NullString
		var adWords sql.NullInt32
		var links sql.NullInt32
		var htmlLink sql.NullString
		var rawHtml sql.NullString
		var searchResults sql.NullInt32
		var timeTaken sql.NullFloat64
		var createdDate sql.NullTime
		var cache sql.NullString

		err := rows.Scan(&id, &keyword, &adWords, &links, &htmlLink, &rawHtml, &searchResults, &timeTaken, &createdDate, &cache)
		if err != nil {
			return nil, err
		}
		detail := model.GetKeywordListResponse{
			Id:            util.GetStringFromSQL(id),
			Keyword:       util.GetStringFromSQL(keyword),
			AdWords:       util.GetIntFromSQL(adWords),
			Links:         util.GetIntFromSQL(links),
			HTMLLink:      util.GetStringFromSQL(htmlLink),
			SearchResults: util.GetIntFromSQL(searchResults),
			Cache:         util.GetStringFromSQL(cache),
			TimeTaken:     util.GetFloatFromSQL(timeTaken),
			RawHTML:       util.GetStringFromSQL(rawHtml),
		}
		if date := util.GetTimeFromSQL(createdDate); date != nil && !date.IsZero() {
			detail.CreatedDate, _ = util.TimestampToStringWithLocation("", date.Unix(), "Asia/Bangkok")
		}
		details = append(details, detail)
	}
	return details, nil
}
