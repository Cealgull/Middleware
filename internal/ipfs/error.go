package ipfs

import (
	"net/http"

	"github.com/Cealgull/Middleware/internal/proto"
)

type IPFSBackendError struct{}

func (e *IPFSBackendError) Status() int {
	return http.StatusOK
}

func (e *IPFSBackendError) Message() *proto.ResponseMessage {
	return &proto.ResponseMessage{
		Code:    "I0001",
		Message: e.Error(),
	}
}

func (e *IPFSBackendError) Error() string {
	return "IPFS: IPFS Storage Backend Error or Timeout."
}

type HeaderMissingError struct{}
type FileMissingError struct{}

func (e *HeaderMissingError) Error() string {
	return "Bad Request: Missing headers in request."
}

func (e *HeaderMissingError) Status() int {
	return http.StatusBadRequest
}

func (e *HeaderMissingError) Message() *proto.ResponseMessage {
	return &proto.ResponseMessage{
		Code:    "B0001",
		Message: e.Error(),
	}
}

func (e *FileMissingError) Status() int {
	return http.StatusBadRequest
}

func (e *FileMissingError) Error() string {
	return "Bad Request: Missing file payload when uploading."
}

func (e *FileMissingError) Message() *proto.ResponseMessage {
	return &proto.ResponseMessage{
		Code:    "B0002",
		Message: e.Error(),
	}
}

var success *proto.Success = &proto.Success{}
var headerMissingError *HeaderMissingError = &HeaderMissingError{}
var fileMissingError *FileMissingError = &FileMissingError{}
var backendError *IPFSBackendError = &IPFSBackendError{}
