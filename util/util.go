package util

import (
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"github.com/go-playground/validator"
	"github.com/labstack/echo/v4"
	"github.com/supanutjarukulgowit/google_search_web_api/common"
	"github.com/supanutjarukulgowit/google_search_web_api/database"
	"github.com/supanutjarukulgowit/google_search_web_api/model"
	"github.com/supanutjarukulgowit/google_search_web_api/static"
)

func ByteToStruct(src []byte, des interface{}) error {
	err := json.Unmarshal(src, des)
	if err != nil {
		return fmt.Errorf("ByteToStruct json Unmarshal error : %v", err)
	}
	return nil
}

func InterfaceToStruct(src, des interface{}) error {
	if src == nil {
		return fmt.Errorf("InterfaceToStruct src is nil")
	}
	byteData, err := json.Marshal(src)
	if err != nil {
		return fmt.Errorf("InterfaceToStruct json Marshal error: %v", err)
	}

	err = json.Unmarshal(byteData, des)
	if err != nil {
		return fmt.Errorf("InterfaceToStruct json Unmarshal error: %v", err)
	}

	return nil
}

func GenError(c echo.Context, errCode, errMsg, devMsg string, httpStatus interface{}) common.TemplateResponse {
	uuid, _ := GetUUID()
	// userID := ""
	statusCode := http.StatusBadRequest
	if v, ok := httpStatus.(int); ok && v != 0 {
		statusCode = v
	}

	templateResponse := common.TemplateResponse{
		HTTPStatus: statusCode,
		Response: common.ResponseObj{
			Status: common.StatusObj{
				Code:        "-1",
				Description: "Failure",
			},
			Date: time.Now().Format("2006-01-02 15:04:05"),
		},
		Error: &common.ErrorObj{
			ErrorID:            uuid,
			Created:            time.Now().Format(time.RFC3339Nano),
			ErrorCode:          errCode,
			MessageToUser:      static.ERROR_DESC[errCode],
			MessageToDeveloper: devMsg,
		},
	}
	return templateResponse
}

func GenResponse(c echo.Context, response interface{}) common.TemplateResponse {
	templateResponse := common.TemplateResponse{
		Response: common.ResponseObj{
			Status: common.StatusObj{
				Code:        "0",
				Description: "Success",
			},
			Date: time.Now().Format("2006-01-02 15:04:05"),
		},
		Data: response,
	}
	return templateResponse
}

func GetUUID() (string, error) {
	out, err := exec.Command("uuidgen").Output()
	if err != nil {
		return "", err
	}

	r := strings.NewReplacer("\n", "", "-", "")
	return r.Replace(string(out)), nil
}

func ValidatorParam(req interface{}) error {
	v := validator.New()
	err := v.Struct(req)
	if err != nil {
		msg := ""
		for _, e := range err.(validator.ValidationErrors) {
			msg += fmt.Sprintf("%v ", e)
		}

		msg = fmt.Sprintf("%s %s", err.Error(), msg)
		return fmt.Errorf(msg)
	}

	return nil
}

func LoadDBConfig(key string, result interface{}, pg *database.PostgreSQL, pgConnection *model.PostgreSQLConnect) error {
	db, err := pg.ConnectPostgreSQLGorm(pgConnection.Host, pgConnection.User, pgConnection.Password, pgConnection.Database, pgConnection.Port)
	if err != nil {
		return err
	}
	config := &model.ConfigurationDb{}
	db.Where("id = ?", key).First(config)
	if config.Id != "" {
		err := json.Unmarshal([]byte(config.Val), &result)
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("cannot load Config of key %s", key)
	}
	return nil
}

// GetStringFromSQL ... get string when query sqldb
func GetStringFromSQL(val sql.NullString) string {
	if val.Valid {
		return val.String
	}

	return ""
}

// GetIntFromSQL ... get int when query sqldb
func GetIntFromSQL(val sql.NullInt32) int {
	if val.Valid {
		return int(val.Int32)
	}

	return 0
}

// GetIntFromSQL ... get int when query sqldb
func GetInt64FromSQL(val sql.NullInt64) int {
	if val.Valid {
		return int(val.Int64)
	}

	return 0
}

// GetTimeFromSQL ... get time when query sqldb
func GetTimeFromSQL(val sql.NullTime) *time.Time {
	if val.Valid {
		strTime, _ := TimestampToStringWithLocation("", val.Time.Unix(), "UTC")
		// loc, _ := time.LoadLocation("Asia/Bangkok")
		t, _ := time.ParseInLocation("2006-01-02 15:04:05", strTime, time.Local)
		return &t
	}

	return nil
}

// GetTimeFromSQL ... get time when query sqldb
func GetFloatFromSQL(val sql.NullFloat64) float64 {
	if val.Valid {
		return val.Float64
	}

	return 0
}

// TimestampToStringWithLocation ... convert string to unix timestamp
func TimestampToStringWithLocation(format string, sec int64, local string) (string, error) {
	if format == "" {
		format = "2006-01-02 15:04:05"
	}

	t := time.Unix(sec, 0)
	loc, err := time.LoadLocation(local)
	if err != nil {
		return "", err
	}

	return t.In(loc).Format(format), nil
}

// TimestampToString ... convert string to unix timestamp
func TimestampToString(format string, sec int64) string {
	if format == "" {
		format = "2006-01-02 15:04:05"
	}

	t := time.Unix(sec, 0)
	return t.Format(format)
}

func ReadUploadCsvFile(file *multipart.FileHeader, columnName []string) ([][]string, string, error) {
	src, err := file.Open()
	if err != nil {
		return nil, static.CANNOT_READ_FILE_ERROR, fmt.Errorf("open file error %s", err.Error())
	}
	defer src.Close()

	csvReader := csv.NewReader(src)
	records, err := csvReader.ReadAll()
	if err != nil {
		return nil, static.CANNOT_READ_FILE_ERROR, fmt.Errorf("read file error %s", err.Error())
	}
	if len(records) == 1 {
		return nil, static.FILE_NO_DATA_ERROR, fmt.Errorf("no data found")
	}

	if len(records[0]) < len(columnName) {
		return nil, static.COLUMN_COUNT_INVALID, fmt.Errorf("column count invalid")
	}

	for index, col := range records[0][0:len(columnName)] {
		if col != columnName[index] {
			return nil, static.COLUMN_NAME_INVALID, fmt.Errorf("column name invalid")
		}
	}
	return records[1:], "", nil
}
