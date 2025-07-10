package store

// Хранилище ошибок
const (
	DefaultError           = "Error"
	DefaultErrorCode       = 400
	InternalSeverErrorCode = 500
	ReadBodyError          = "Failed to read request body"
	LargeBodyError         = "Request body too large"
	ConnectionError        = "Connection error"
	BadRequestError        = "Bad request"
)
