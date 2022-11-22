package model

import "time"

type DownloadTemplateResponse struct {
	Base64 string `json:"base64"`
}

type GoogleSearchApiresponse struct {
	Keyword           string
	SearchMetadata    SearchMetadata `mapstructure:"search_metadata"`
	SearchInformation struct {
		TimeTakenDisplayed float64 `mapstructure:"time_taken_displayed"`
		TotalResults       int     `mapstructure:"total_results"`
	} `mapstructure:"search_information"`
	Ads []struct {
		Position int `json:"position"`
	} `mapstructure:"ads"`
	TotalLinks int
}

type SearchMetadata struct {
	GoogleUrl string `mapstructure:"google_url"`
	HtmlCode  string
}

type Sitelinks struct {
	Expanded []Expanded `mapstructure:"expanded"`
}

type Expanded struct {
	Title   string `mapstructure:"title"`
	Link    string `mapstructure:"link"`
	Snippet string `mapstructure:"snippet"`
}

type GoogleSearchApiDb struct {
	Id          string
	SearchId    string
	UserId      string
	FileName    string
	CreatedDate time.Time
}

type GoogleSearchApiDetailDb struct {
	Id            string
	Keyword       string
	SearchId      string
	AdWords       int
	Links         int
	HTMLLink      string
	RawHTML       []byte
	SearchResults int
	TimeTaken     float64
	Cache         string
	CreatedDate   time.Time
	UserId        string
}

type GetKeywordListResponse struct {
	Id            string  `json:"id"`
	Keyword       string  `json:"keyword"`
	AdWords       int     `json:"ads_words"`
	Links         int     `json:"link"`
	HTMLLink      string  `json:"html_link"`
	RawHTML       string  `json:"raw_html"`
	SearchResults int     `json:"search_results"`
	TimeTaken     float64 `json:"time_taken"`
	Cache         string  `json:"cache_link"`
	CreatedDate   string  `json:"created_date"`
}
