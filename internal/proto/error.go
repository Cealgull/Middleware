package proto

import "net/http"

type MiddlewareError interface {
	error
	Message() *ResponseMessage
	Status() int
}

type Success struct{}

func (e *Success) Error() string {
	return "OK"
}

func (e *Success) Status() int {
	return http.StatusOK
}

func (e *Success) Message() *ResponseMessage {
	return &ResponseMessage{
		Code:    "N0001",
		Message: e.Error(),
	}
}
