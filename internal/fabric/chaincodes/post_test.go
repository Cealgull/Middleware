package chaincodes

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/Cealgull/Middleware/internal/models"
	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"

	"github.com/Cealgull/Middleware/internal/fabric/common/mocks"
	fabricmock "github.com/Cealgull/Middleware/internal/fabric/common/mocks"
	ipfsmock "github.com/Cealgull/Middleware/internal/ipfs/mocks"
)

func preparePostData(t *testing.T) *gorm.DB {

	user := &User{
		Username: "Admin",
		Wallet:   "0x123456789",
		Avatar:   "null",
	}

	topic := &Topic{
		Title:   "Hello world",
		Hash:    "topic",
		Creator: user,
	}

	db := newSqliteDB()

	assert.NoError(t, db.Create(&user).Error)

	assert.NoError(t, db.Create(&topic).Error)

	posts := []*Post{
		{
			Hash:     "post1",
			Content:  "Hello world!",
			Creator:  user,
			BelongTo: topic,
		},
		{
			Hash:     "post2",
			Content:  "Hello world!",
			Creator:  user,
			BelongTo: topic,
		},
	}

	assert.NoError(t, db.Create(&posts).Error)

	return db
}

func TestInvokeCreatePost(t *testing.T) {

	type PostRequest struct {
		Content  string   `json:"content"`
		Images   []string `json:"images"`
		ReplyTo  string   `json:"replyTo"`
		BelongTo string   `json:"belongTo"`
	}

	payload := PostRequest{
		Content:  "Hello world",
		Images:   []string{},
		ReplyTo:  "post2",
		BelongTo: "topic",
	}

	storage := ipfsmock.NewMockIPFSStorage(t)
	storage.EXPECT().Version().Return("abcd", "abcd", nil).Once()

	ipfs := NewMockIPFSManager(storage)

	contract := fabricmock.NewMockContract()

	db := preparePostData(t)

	createPost := invokeCreatePost(logger, ipfs, db)

	t.Run("Creating Post With Unmarshal Error", func(t *testing.T) {

		req := httptest.NewRequest(http.MethodPost, "/api/post/invoke/CreatePost", bytes.NewReader([]byte{1, 2, 3}))
		rec := httptest.NewRecorder()

		c := server.NewContext(req, rec)
		c = newMockSignedContext(c)

		err := createPost(contract, c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("Creating Post With BelongTo Error", func(t *testing.T) {
		payload.BelongTo = "abcd"

		req := httptest.NewRequest(http.MethodPost, "/api/post/invoke/CreatePost", newJsonRequest(&payload))
		req.Header.Add(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		c := server.NewContext(req, rec)
		c = newMockSignedContext(c)

		err := createPost(contract, c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)

		payload.BelongTo = "topic"
	})

	t.Run("Creating Post With ReplyTo Error", func(t *testing.T) {
		payload.ReplyTo = "abcd"

		req := httptest.NewRequest(http.MethodPost, "/api/post/invoke/CreatePost", newJsonRequest(&payload))
		req.Header.Add(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		c := server.NewContext(req, rec)
		c = newMockSignedContext(c)

		err := createPost(contract, c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)

		payload.ReplyTo = "post2"
	})

	t.Run("Creating Post With IPFS Failure", func(t *testing.T) {

		storage.On("Add", mock.Anything).Return("", errors.New("hello world")).Once()

		req := httptest.NewRequest(http.MethodPost, "/api/post/invoke/CreatePost", newJsonRequest(&payload))
		req.Header.Add(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		c := server.NewContext(req, rec)
		c = newMockSignedContext(c)

		err := createPost(contract, c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)

	})

	t.Run("Creating Post With Base64DecodeError", func(t *testing.T) {

		payload.Images = []string{"base64Error&*"}
		storage.On("Add", mock.Anything).Return("base64", nil).Once()

		req := httptest.NewRequest(http.MethodPost, "/api/post/invoke/CreatePost", newJsonRequest(&payload))
		req.Header.Add(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		c := server.NewContext(req, rec)
		c = newMockSignedContext(c)

		err := createPost(contract, c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)

	})

	t.Run("Creating Post With IPFS Error on Uploading Images", func(t *testing.T) {

		payload.Images = []string{base64.StdEncoding.EncodeToString([]byte("base64Error&*"))}

		storage.On("Add", mock.Anything).Return("base64", nil).Once()
		storage.On("Add", mock.Anything).Return("", errors.New("Hello world")).Once()

		req := httptest.NewRequest(http.MethodPost, "/api/post/invoke/CreatePost", newJsonRequest(&payload))
		req.Header.Add(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		c := server.NewContext(req, rec)
		c = newMockSignedContext(c)

		err := createPost(contract, c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})

	t.Run("Creating Post With Chaincode Network Failure", func(t *testing.T) {

		payload.Images = []string{base64.StdEncoding.EncodeToString([]byte("base64Error&*"))}

		storage.On("Add", mock.Anything).Return("base64", nil)

		req := httptest.NewRequest(http.MethodPost, "/api/post/invoke/CreatePost", newJsonRequest(&payload))
		req.Header.Add(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		c := server.NewContext(req, rec)
		c = newMockSignedContext(c)

		contract.On("Submit", "CreatePost", mock.Anything).Return([]byte(nil), errors.New("Hello world")).Once()
		err := createPost(contract, c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)

	})

	t.Run("Creating Post With Success", func(t *testing.T) {

		payload.Images = []string{base64.StdEncoding.EncodeToString([]byte("base64Error&*"))}

		storage.On("Add", mock.Anything).Return("base64", nil)

		req := httptest.NewRequest(http.MethodPost, "/api/post/invoke/CreatePost", newJsonRequest(&payload))
		req.Header.Add(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		c := server.NewContext(req, rec)
		c = newMockSignedContext(c)

		contract.On("Submit", "CreatePost", mock.Anything).Return([]byte(nil), nil).Once()
		err := createPost(contract, c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
	})
}

func TestCreatePostCallback(t *testing.T) {

	postBlock := PostBlock{
		CID:      "abcd",
		Hash:     "abcd",
		Creator:  "0x123456789",
		ReplyTo:  "post2",
		BelongTo: "topic",
		Assets:   []string{"abcd"},
	}

	storage := ipfsmock.NewMockIPFSStorage(t)
	storage.EXPECT().Version().Return("abcd", "abcd", nil).Once()

	ipfs := NewMockIPFSManager(storage)

	db := preparePostData(t)

	createPost := createPostCallback(logger, ipfs, db)

	t.Run("Create Post Callback With IPFS failure", func(t *testing.T) {

		storage.EXPECT().Cat(postBlock.CID).Return(io.ReadCloser(nil), errors.New("hello world")).Once()
		b, _ := json.Marshal(&postBlock)

		err := createPost(b)
		assert.Error(t, err)

	})

	reader := io.NopCloser(bytes.NewReader([]byte("document")))

	t.Run("Creating Post Callback with replyTo error", func(t *testing.T) {

		storage.EXPECT().Cat(postBlock.CID).Return(reader, nil)
		postBlock.ReplyTo = "unknown"
		b, _ := json.Marshal(&postBlock)

		err := createPost(b)
		assert.Error(t, err)
		postBlock.ReplyTo = "post2"
	})

	t.Run("Creating Post Callback with belongTo error", func(t *testing.T) {

		storage.EXPECT().Cat(postBlock.CID).Return(reader, nil)
		postBlock.BelongTo = "unknown"
		b, _ := json.Marshal(&postBlock)

		err := createPost(b)
		assert.Error(t, err)
		postBlock.BelongTo = "topic"
	})

	t.Run("Creating Post Callback with user not found", func(t *testing.T) {

		storage.EXPECT().Cat(postBlock.CID).Return(reader, nil)
		postBlock.Creator = "unknown"
		b, _ := json.Marshal(&postBlock)

		err := createPost(b)
		assert.Error(t, err)
		postBlock.Creator = "0x123456789"
	})

	t.Run("Creating Post Callback with success", func(t *testing.T) {

		storage.EXPECT().Cat(postBlock.CID).Return(reader, nil)
		b, _ := json.Marshal(&postBlock)

		err := createPost(b)
		assert.NoError(t, err)
	})

}

func TestInvokeUpdatePost(t *testing.T) {
	type UpdatePostRequest struct {
		Hash    string   `json:"hash"`
		Content string   `json:"content"`
		Images  []string `json:"assets"`
		Type    string   `json:"type"`
	}

	payload := UpdatePostRequest{
		Hash:    "post1",
		Content: "Hello world",
		Images:  []string{},
		Type:    "post",
	}

	storage := ipfsmock.NewMockIPFSStorage(t)
	storage.EXPECT().Version().Return("abcd", "abcd", nil).Once()

	ipfs := NewMockIPFSManager(storage)

	contract := fabricmock.NewMockContract()

	db := preparePostData(t)

	updatePost := invokeUpdatePost(logger, ipfs, db)
	t.Run("Updating Post With Unmarshal Error", func(t *testing.T) {

		req := httptest.NewRequest(http.MethodPost, "/api/post/invoke/UpdatePost", bytes.NewReader([]byte{1, 2, 3}))
		rec := httptest.NewRecorder()

		c := server.NewContext(req, rec)
		c = newMockSignedContext(c)

		err := updatePost(contract, c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("Updating Post With Hash Error", func(t *testing.T) {
		payload.Hash = "a111"
		req := httptest.NewRequest(http.MethodPost, "/api/post/invoke/UpdatePost", newJsonRequest(&payload))
		req.Header.Add(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		c := server.NewContext(req, rec)
		c = newMockSignedContext(c)

		err := updatePost(contract, c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		payload.Hash = "post1"
	})

	t.Run("Updating Post With IPFS Failure", func(t *testing.T) {

		storage.On("Add", mock.Anything).Return("", errors.New("hello world")).Once()

		req := httptest.NewRequest(http.MethodPost, "/api/post/invoke/UpdatePost", newJsonRequest(&payload))
		req.Header.Add(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		c := server.NewContext(req, rec)
		c = newMockSignedContext(c)

		err := updatePost(contract, c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)

	})

	t.Run("Updating Post With Base64DecodeError", func(t *testing.T) {

		payload.Images = []string{"base64Error&*"}
		storage.On("Add", mock.Anything).Return("base64", nil).Once()

		req := httptest.NewRequest(http.MethodPost, "/api/post/invoke/UpdatePost", newJsonRequest(&payload))
		req.Header.Add(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		c := server.NewContext(req, rec)
		c = newMockSignedContext(c)

		err := updatePost(contract, c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)

	})

	t.Run("Updating Post With IPFS Error on Uploading Images", func(t *testing.T) {

		payload.Images = []string{base64.StdEncoding.EncodeToString([]byte("base64Error&*"))}

		storage.On("Add", mock.Anything).Return("base64", nil).Once()
		storage.On("Add", mock.Anything).Return("", errors.New("Hello world")).Once()

		req := httptest.NewRequest(http.MethodPost, "/api/post/invoke/UpdatePost", newJsonRequest(&payload))
		req.Header.Add(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		c := server.NewContext(req, rec)
		c = newMockSignedContext(c)

		err := updatePost(contract, c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})

	t.Run("Updating Post With Chaincode Network Failure", func(t *testing.T) {

		payload.Images = []string{base64.StdEncoding.EncodeToString([]byte("base64Error&*"))}

		storage.On("Add", mock.Anything).Return("base64", nil)

		req := httptest.NewRequest(http.MethodPost, "/api/post/invoke/UpdatePost", newJsonRequest(&payload))
		req.Header.Add(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		c := server.NewContext(req, rec)
		c = newMockSignedContext(c)

		contract.On("Submit", "UpdatePost", mock.Anything).Return([]byte(nil), errors.New("Hello world")).Once()
		err := updatePost(contract, c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)

	})

	t.Run("Updating Post With Success", func(t *testing.T) {

		payload.Images = []string{base64.StdEncoding.EncodeToString([]byte("base64Error&*"))}

		storage.On("Add", mock.Anything).Return("base64", nil)

		req := httptest.NewRequest(http.MethodPost, "/api/post/invoke/UpdatePost", newJsonRequest(&payload))
		req.Header.Add(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		c := server.NewContext(req, rec)
		c = newMockSignedContext(c)

		contract.On("Submit", "UpdatePost", mock.Anything).Return([]byte(nil), nil).Once()
		err := updatePost(contract, c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
	})

}

func TestUpdatePostCallback(t *testing.T) {

	postBlock := PostBlock{
		CID:      "defg",
		Hash:     "post1",
		Creator:  "0x123456789",
		ReplyTo:  "post2",
		BelongTo: "topic",
		Assets:   []string{"abcd"},
	}

	storage := ipfsmock.NewMockIPFSStorage(t)
	storage.EXPECT().Version().Return("abcd", "abcd", nil).Once()

	ipfs := NewMockIPFSManager(storage)

	db := preparePostData(t)

	updatePost := updatePostCallback(logger, ipfs, db)

	t.Run("Update Post Callback With IPFS failure", func(t *testing.T) {

		storage.EXPECT().Cat(postBlock.CID).Return(io.ReadCloser(nil), errors.New("hello world")).Once()
		b, _ := json.Marshal(&postBlock)

		err := updatePost(b)
		assert.Error(t, err)

	})

	reader := io.NopCloser(bytes.NewReader([]byte("document")))

	t.Run("Creating Post Callback with success", func(t *testing.T) {

		storage.EXPECT().Cat(postBlock.CID).Return(reader, nil)

		b, _ := json.Marshal(&postBlock)

		err := updatePost(b)
		assert.NoError(t, err)
	})

	t.Run("Creating Post Callback with hash not found", func(t *testing.T) {
		postBlock.Hash = "unknown"

		b, _ := json.Marshal(&postBlock)

		err := updatePost(b)
		assert.Error(t, err)
	})

}

func TestQueryPostsList(t *testing.T) {
	type QueryRequest struct {
		PageOrdinal int    `json:"pageOrdinal"`
		PageSize    int    `json:"pageSize"`
		BelongTo    string `json:"belongTo"`
		Creator     string `json:"creator"`
	}

	payload := QueryRequest{
		PageOrdinal: 1,
		PageSize:    10,
		BelongTo:    "topic",
		Creator:     "0x123456789",
	}

	t.Run("Query Post List With Unmarshal Error", func(t *testing.T) {

		db := newSqliteDB()

		query := queryPostsList(logger, db)

		req := httptest.NewRequest(http.MethodPost, "/api/post/query/list", bytes.NewReader([]byte{1, 2, 3}))
		rec := httptest.NewRecorder()
		c := server.NewContext(req, rec)

		err := query(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("Querying List With Illegal PageOrdinal", func(t *testing.T) {

		db := newSqliteDB()

		query := queryPostsList(logger, db)

		payload.PageOrdinal = -1
		req := httptest.NewRequest(http.MethodPost, "/api/post/query/list", newJsonRequest(&payload))
		req.Header.Add(echo.HeaderContentType, echo.MIMEApplicationJSON)

		rec := httptest.NewRecorder()
		c := server.NewContext(req, rec)

		err := query(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
		payload.PageOrdinal = 1
	})

	t.Run("Querying List With Success", func(t *testing.T) {

		db := newSqliteDB()

		query := queryPostsList(logger, db)

		req := httptest.NewRequest(http.MethodPost, "/api/post/query/list", newJsonRequest(&payload))
		req.Header.Add(echo.HeaderContentType, echo.MIMEApplicationJSON)

		rec := httptest.NewRecorder()
		c := server.NewContext(req, rec)

		err := query(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
	})
}

func TestNewPostChaincodeMiddleware(t *testing.T) {

	network := mocks.NewMockNetwork(t)

	storage := ipfsmock.NewMockIPFSStorage(t)
	storage.EXPECT().Version().Return("abcd", "abcd", nil).Once()

	ipfs := NewMockIPFSManager(storage)

	db := newSqliteDB()
	network.EXPECT().GetContract("post").Return(&client.Contract{}).Once()
	var _ = NewPostChaincodeMiddleware(logger, network, ipfs, db)

}
