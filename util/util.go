package util

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"github.com/go-playground/validator"
	"github.com/labstack/echo/v4"
	"github.com/supanutjarukulgowit/google_search_web_api/common"
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
