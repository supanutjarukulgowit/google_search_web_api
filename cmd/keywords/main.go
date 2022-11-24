package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/sirupsen/logrus"
	"github.com/supanutjarukulgowit/google_search_web_api/configuration"
	"github.com/supanutjarukulgowit/google_search_web_api/di"
	"github.com/supanutjarukulgowit/google_search_web_api/handler"
	"github.com/supanutjarukulgowit/google_search_web_api/interceptor"
)

var (
	configPath = flag.String("config", "", "")
	version    = flag.String("version", "unknown", "")
	envFile    = flag.String("env", "", "")
	Log        = logrus.New()
)

func main() {
	//Echo instance
	e := echo.New()

	//Receive args
	flag.Parse()

	var err error
	if *envFile != "" {
		err = godotenv.Load(*envFile)
		if err != nil {
			Log.Fatal("Load env file error ! %s %s", err.Error(), *envFile)
		}
	}

	if os.Getenv("API_VERSION") != "" {
		*version = os.Getenv("API_VERSION")
	}

	config, err := configuration.LoadConfigFile(*configPath)
	if err != nil {
		Log.Fatal("LoadConfigFile error : %s", err.Error())
	}

	fmt.Println("API_VERSION: ", os.Getenv("API_VERSION"))
	fmt.Println("BINARY_NAME", os.Getenv("BINARY_NAME"))
	fmt.Println("CONFIG_FILE", os.Getenv("CONFIG_FILE"))
	fmt.Println("PORT", os.Getenv("PORT"))

	e.Use(middleware.Recover())

	//CORS
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{echo.GET, echo.HEAD, echo.PUT, echo.PATCH, echo.POST, echo.DELETE},
		AllowCredentials: true,
	}))
	di.Init(config)
	h, err := handler.NewKeywordsHandler(config.PostgreSQL, config.GoogleSearchApiKey)
	if err != nil {
		Log.Fatal("NewKeywordsHandler error : %s", err.Error())
	}
	g := e.Group("")
	g.Use(interceptor.ValidateToken())
	e.GET("/Health", Health)
	g.GET("/api/keywords/download/template", h.DownloadTemplate)
	g.POST("/api/keywords/upload/file", h.UploadFile)
	g.GET("/api/keywords/list", h.GetKeywordList)
	g.POST("/api/keywords/search", h.GetSearchKeyword)

	e.Logger.Fatal(e.Start(":" + os.Getenv("PORT")))
}

func Health(c echo.Context) error {
	type Message struct {
		Status  string `json:"status"`
		Version string `json:"version"`
	}
	return c.JSON(http.StatusOK, Message{Status: "OK", Version: *version})
}
