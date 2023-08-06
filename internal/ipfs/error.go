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

type UploadHeaderMissingError struct{}

func (e *UploadHeaderMissingError) Error() string {
	return "Bad Request: Missing headers in request."
}

func (e *UploadHeaderMissingError) Status() int {
	return http.StatusBadRequest
}

func (e *UploadHeaderMissingError) Message() *proto.ResponseMessage {
	return &proto.ResponseMessage{
		Code:    "B0001",
		Message: e.Error(),
	}
}

type UploadFileMissingError struct{}

func (e *UploadFileMissingError) Status() int {
	return http.StatusBadRequest
}

func (e *UploadFileMissingError) Error() string {
	return "Bad Request: Missing file payload when uploading."
}

func (e *UploadFileMissingError) Message() *proto.ResponseMessage {
	return &proto.ResponseMessage{
		Code:    "B0002",
		Message: e.Error(),
	}
}

type IPFSFileNotFoundError struct{}

func (e *IPFSFileNotFoundError) Status() int {
  return http.StatusNotFound
}

func (e *IPFSFileNotFoundError) Error() string {
  return "Not Found: File not found."
}

func (e *IPFSFileNotFoundError) Message() *proto.ResponseMessage {
  return &proto.ResponseMessage{
    Code:    "B0003",
    Message: e.Error(),
  }
}


var success *proto.Success = &proto.Success{}
var uploadHeaderMissingError *UploadHeaderMissingError = &UploadHeaderMissingError{}
var uploadFileMissingError *UploadFileMissingError = &UploadFileMissingError{}
var ipfsBackendError *IPFSBackendError = &IPFSBackendError{}
