package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/supanutjarukulgowit/google_search_web_api/model"
	"github.com/supanutjarukulgowit/google_search_web_api/util"
	"gorm.io/gorm"
)

func GetSearchedKeyword(db *gorm.DB, keywords []string) ([]string, error) {
	keywordFound := make([]string, 0)
	param := ""
	for _, k := range keywords {
		if param != "" {
			param += fmt.Sprintf(", '%s'", k)
		} else {
			param += fmt.Sprintf("'%s'", k)
		}
	}
	// Raw SQL
	query := fmt.Sprintf(`select distinct keyword from google_search_api_detail_dbs
	where keyword in (%s) and status = 'success'`, param)
	rows, err := db.Raw(query).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var keyword sql.NullString
		err := rows.Scan(&keyword)
		if err != nil {
			return nil, err
		}
		keywordFound = append(keywordFound, util.GetStringFromSQL(keyword))
	}
	return keywordFound, nil
}

func SaveSearchData(db *gorm.DB, fileName, userID, searchID string) error {
	id, _ := util.GetUUID()
	search := model.GoogleSearchApiDb{
		Id:          id,
		SearchId:    searchID,
		UserId:      userID,
		FileName:    fileName + "_" + time.Now().Format("2006-01-02_15:04:05"),
		CreatedDate: time.Now(),
	}
	db.Create(&search)
	return nil
}

func GetFoundKeywords(db *gorm.DB, foundKeywords map[string]string, userID, searchID string) ([]model.GoogleSearchApiDetailDb, error) {
	details := []model.GoogleSearchApiDetailDb{}
	for _, v := range foundKeywords {
		rows, err := db.Raw(`select keyword, ad_words, links,
		html_link, raw_html, search_results, time_taken, created_date, cache from google_search_api_detail_dbs
		where keyword = ? and user_id = ? ORDER BY created_date desc limit 1`, v, userID).Rows()
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		if rows.Next() {
			var keyword sql.NullString
			var adWords sql.NullInt32
			var links sql.NullInt32
			var htmlLink sql.NullString
			var rawHtml sql.NullString
			var searchResults sql.NullInt64
			var timeTaken sql.NullFloat64
			var createdDate sql.NullTime
			var cache sql.NullString

			err := rows.Scan(&keyword, &adWords, &links, &htmlLink, &rawHtml, &searchResults, &timeTaken, &createdDate, &cache)
			if err != nil {
				return nil, err
			}
			uuid, _ := util.GetUUID()
			loc, _ := time.LoadLocation("Asia/Bangkok")
			detail := model.GoogleSearchApiDetailDb{
				Id:            uuid,
				Keyword:       util.GetStringFromSQL(keyword),
				AdWords:       util.GetIntFromSQL(adWords),
				Links:         util.GetIntFromSQL(links),
				HTMLLink:      util.GetStringFromSQL(htmlLink),
				SearchResults: util.GetInt64FromSQL(searchResults),
				Cache:         util.GetStringFromSQL(cache),
				TimeTaken:     util.GetFloatFromSQL(timeTaken),
				RawHTML:       []byte(util.GetStringFromSQL(rawHtml)),
				CreatedDate:   time.Now().In(loc),
				UserId:        userID,
				SearchId:      searchID,
				Status:        "success",
			}
			details = append(details, detail)
		}
	}

	return details, nil
}

func UpdateSearchDataDetail(keywordsMap map[string]*model.GoogleSearchApiDetailDb, db *gorm.DB) error {
	for _, data := range keywordsMap {
		update := map[string]interface{}{
			"ad_words":       data.AdWords,
			"links":          data.Links,
			"html_link":      data.HTMLLink,
			"raw_html":       data.RawHTML,
			"search_results": data.SearchResults,
			"time_taken":     data.TimeTaken,
			"cache":          data.Cache,
			"status":         data.Status,
			"err_msg":        data.ErrMsg,
		}
		r := db.Model(data).Where("user_id = ? and id = ? and keyword = ?", data.UserId, data.Id, data.Keyword).Updates(update)
		if r.Error != nil {
			return r.Error
		}
	}
	return nil
}

func InsertSearchDataDetail(db *gorm.DB, result []model.GoogleSearchApiresponse, userID, searchID string) error {
	searchDetails := make([]model.GoogleSearchApiDetailDb, 0)
	for _, r := range result {
		detailID, _ := util.GetUUID()
		loc, _ := time.LoadLocation("Asia/Bangkok")
		detail := model.GoogleSearchApiDetailDb{
			Id:            detailID,
			SearchId:      searchID,
			Keyword:       r.Keyword,
			AdWords:       len(r.Ads),
			Links:         r.TotalLinks,
			HTMLLink:      r.SearchMetadata.GoogleUrl,
			SearchResults: r.SearchInformation.TotalResults,
			TimeTaken:     r.SearchInformation.TimeTakenDisplayed,
			CreatedDate:   time.Now().In(loc),
			UserId:        userID,
			RawHTML:       []byte(r.SearchMetadata.HtmlCode),
		}
		searchDetails = append(searchDetails, detail)
	}
	db.CreateInBatches(&searchDetails, 50)
	return nil
}
