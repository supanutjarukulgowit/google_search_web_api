package common

type TemplateResponse struct {
	Response   ResponseObj `json:"response"`
	Data       interface{} `json:"data"`
	Error      *ErrorObj   `json:"error"`
	HTTPStatus int         `json:"-"`
}

type ResponseObj struct {
	Date   string    `json:"date"`
	Status StatusObj `json:"status"`
}

type StatusObj struct {
	Code        string `json:"code"`
	Description string `json:"description"`
}

type ErrorObj struct {
	ErrorID            string `json:"errorId"`
	ErrorCode          string `json:"code"`
	MessageToDeveloper string `json:"messageToDeveloper"`
	MessageToUser      string `json:"messageToUser"`
	Created            string `json:"created"`
}
