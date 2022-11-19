package model

type PostgreSQLConnect struct {
	Port     int    `json:"port"`
	Database string `json:"database"`
	User     string `json:"user"`
	Password string `json:"password"`
	Host     string `json:"host"`
}

type ConfigurationDb struct {
	Id  string `json:"id"`
	Val string `json:"val"`
}
