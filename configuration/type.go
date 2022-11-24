package configuration

type Configuration struct {
	PostgreSQL         interface{} `json:"postgreSQL"`
	GoogleSearchApiKey string      `json:"google_search_api_key"`
}
