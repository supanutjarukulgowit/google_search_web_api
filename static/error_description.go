package static

var (
	ERROR_DESC = map[string]string{
		INVALID_PARAMS:          `Invalid parameters`,
		INTERNAL_SERVER_ERROR:   `Internal server error`,
		USER_AUTHEN_ERROR:       `User Authen error`,
		USER_NOT_FOUND:          `User not found`,
		USER_WRONG_PASSWORD:     `Wrong password`,
		DOWNLOAD_TEMPLATE_ERROR: `Download template error`,
		CANNOT_READ_FILE_ERROR:  `Cannot read the file`,
		FILE_NO_DATA_ERROR:      `File has no data. keywords must be more than 1`,
		COLUMN_COUNT_INVALID:    `Column count number invalid`,
		COLUMN_NAME_INVALID:     `Column name invalid`,
		CANNOT_LOAD_CONFIG:      `Cannot load config from db`,
		ERROR_AMOUNT_OF_SEARCH:  `Amount of keywords error keywords must be less than 100`,
		GET_KEYWORD_ERROR:       `Get keywords error`,
		UPLOAD_TEMPLATE_ERROR:   `Upload template error`,
		USER_ALREADY_SIGN_UP:    `This username is already taken`,
		UN_AUTH_ERROR:           `Token invalid`,
	}
)
