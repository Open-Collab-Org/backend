package dtos

type ErrorDto struct {
	ErrorCode    string      `json:"errorCode"`
	ErrorDetails interface{} `json:"errorDetails"`
}
