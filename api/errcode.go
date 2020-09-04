package api

//Code 组装Code
type Code struct {
	Code    int64  `json:"code"`
	Message string `json:"message"`
}

var (
	errAPINotDone   = Code{999, "API Not Done"}
	errOK           = Code{0, "OK"}
	errClientError  = Code{400, "Bad Request"}
	errNotfound     = Code{404, "Not found"}
	errUnauthorized = Code{401, "Unauthorized"}
	errNoPermission = Code{403, "No Permission"}
	errServerError  = Code{500, "Server Error"}
	errNotVIPError  = Code{501, "Need VIP"}
)

func serverError(err error) (code Code) {
	code = errServerError
	code.Message = err.Error()
	return code
}

func clientError(err error) (code Code) {
	code = errClientError
	code.Message = err.Error()
	return code
}
