package ipfs

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	. "github.com/Cealgull/Middleware/internal/ipfs/mocks"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

var server *echo.Echo

func TestMain(m *testing.M) {
	server = echo.New()
	var _ = m.Run()
}

func newMockIPFSManager(t *testing.T) (*MockIPFSStorage, *IPFSManager) {

	logger, _ := zap.NewProduction()
	s := NewMockIPFSStorage(t)
	s.EXPECT().Version().Return("0.22.0", "ipfs", nil).Once()
	mgr, err := NewIPFSManager(logger, WithIPFSStorage(s))
	assert.NotNil(t, mgr)
	assert.NoError(t, err)

	return s, mgr
}

func TestNewIPFSManager(t *testing.T) {
	var err error

	logger, _ := zap.NewProduction()
	testurl := "localhost"
	testport := 5001
	mgr, err := NewIPFSManager(logger, WithUrl(testurl, testport))
	assert.Nil(t, mgr)
	assert.Error(t, err)

	var _, _ = newMockIPFSManager(t)
}

func TestIPFSPut(t *testing.T) {

	t.Run("Normal", func(t *testing.T) {
		s, mgr := newMockIPFSManager(t)
		payload := strings.NewReader("test")
		s.EXPECT().Add(payload).Return("QmZv", nil).Once()
		cid, err := mgr.Put(payload)
		assert.Equal(t, "QmZv", cid)
		assert.NoError(t, err)
	})

	t.Run("Backend Error", func(t *testing.T) {
		s, mgr := newMockIPFSManager(t)
		payload := strings.NewReader("test")
		s.EXPECT().Add(payload).Return("", errors.New("hello world")).Once()
		cid, err := mgr.Put(payload)
		assert.Equal(t, "", cid)
		assert.IsType(t, ipfsBackendError, err)
		var _ = err.Status()
		var _ = err.Message()
	})

}

func TestIPFSCat(t *testing.T) {

	t.Run("OK", func(t *testing.T) {
		s, mgr := newMockIPFSManager(t)
		cid := "QmZv"
		s.EXPECT().Cat(cid).Return(io.NopCloser(strings.NewReader("hello world")), nil).Once()
		payload, err := mgr.Cat(cid)
		assert.NotNil(t, payload)
		assert.NoError(t, err)
	})

	t.Run("Storage File Not Found", func(t *testing.T) {
		s, mgr := newMockIPFSManager(t)
		cid := "QmZv"
		s.EXPECT().Cat(cid).Return(nil, errors.New("hello world")).Once()
		payload, err := mgr.Cat(cid)
		assert.Nil(t, payload)
		assert.IsType(t, &StorageFileNotFoundError{}, err)
		var _ = err.Status()
		var _ = err.Message()
	})

}

func TestUpload(t *testing.T) {

	t.Run("Payload With Unmarshal Error", func(t *testing.T) {

		_, mgr := newMockIPFSManager(t)

		req := httptest.NewRequest(http.MethodPost, "/api/upload", strings.NewReader("sadas"))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		c := server.NewContext(req, rec)
		err := mgr.upload(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, c.Response().Status)

	})

	t.Run("Payload With Base64 Encoding Error", func(t *testing.T) {

		_, mgr := newMockIPFSManager(t)

		req := httptest.NewRequest(http.MethodPost, "/api/upload", strings.NewReader(`{"payload":"error"}`))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		c := server.NewContext(req, rec)
		err := mgr.upload(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, c.Response().Status)

	})

	t.Run("Payload with ipfs error", func(t *testing.T) {

		s, mgr := newMockIPFSManager(t)

		payload := strings.NewReader(`{"payload":"aGVsbG8gd29ybGQ="}`)
		s.Mock.On("Add", mock.Anything).Return("", errors.New("helloworld")).Once()

		req := httptest.NewRequest(http.MethodPost, "/api/upload", payload)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		c := server.NewContext(req, rec)
		var _ = mgr.upload(c)
		assert.Equal(t, http.StatusInternalServerError, c.Response().Status)
	})

	t.Run("Payload success", func(t *testing.T) {

		s, mgr := newMockIPFSManager(t)

		payload := strings.NewReader(`{"payload":"aGVsbG8gd29ybGQ="}`)
		s.Mock.On("Add", mock.Anything).Return("base64", nil).Once()

		req := httptest.NewRequest(http.MethodPost, "/api/upload", payload)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		c := server.NewContext(req, rec)
		var _ = mgr.upload(c)
		assert.Equal(t, http.StatusOK, c.Response().Status)
	})

	t.Run("Content-type missing", func(t *testing.T) {

		_, mgr := newMockIPFSManager(t)

		req := httptest.NewRequest(http.MethodPost, "/api/upload", nil)
		rec := httptest.NewRecorder()

		c := server.NewContext(req, rec)
		err := mgr.upload(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, c.Response().Status)
	})

	t.Run("Payload multipartfile missing error", func(t *testing.T) {

		_, mgr := newMockIPFSManager(t)
		req := httptest.NewRequest(http.MethodPost, "/api/upload", nil)
		rec := httptest.NewRecorder()

		c := server.NewContext(req, rec)
		err := mgr.upload(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, c.Response().Status)
	})
}

func TestRegister(t *testing.T) {
	_, mgr := newMockIPFSManager(t)
	var _ = mgr.Register(server)
}
