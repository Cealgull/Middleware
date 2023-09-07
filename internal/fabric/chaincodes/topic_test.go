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

	"github.com/Cealgull/Middleware/internal/ipfs"
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

func NewMockIPFSManager(storage *ipfsmock.MockIPFSStorage) *ipfs.IPFSManager {
	mgr, _ := ipfs.NewIPFSManager(logger, ipfs.WithIPFSStorage(storage))
	return mgr
}

func prepareTopicData(t *testing.T) *gorm.DB {

	user := &User{
		Username: "Admin",
		Wallet:   "0x123456789",
		Avatar:   "null",
	}

	tags := []*Tag{
		{Name: "Genshin Impact", Creator: user, Description: "Genshin Impact"},
		{Name: "Honkai Impact", Creator: user, Description: "Honkai Impact"},
	}

	topic := &Topic{
		Hash:          "topic1",
		Title:         "This is a testing topic",
		CreatorWallet: "0x123456789",
		Creator:       user,
		Content:       "Hello world",
	}

	db := newSqliteDB()

	assert.NoError(t, db.Create(&user).Error)
	assert.NoError(t, db.Create(&tags).Error)
	assert.NoError(t, db.Create(&CategoryGroup{
		Name:  "Games",
		Color: "123456",
		Categories: []*Category{{
			Name:  "Mihoyo",
			Color: "123456",
		},
		},
	}).Error)
	assert.NoError(t, db.Create(&topic).Error)

	return db
}

func TestInvokeCreateTopic(t *testing.T) {

	type TopicRequest struct {
		Content  string   `json:"content"`
		Images   []string `json:"images"`
		Title    string   `json:"title"`
		Category string   `json:"category"`
		Tags     []string `json:"tags"`
	}

	payload := TopicRequest{
		Content:  "Hello world",
		Images:   []string{},
		Title:    "This is a testing topic",
		Category: "Mihoyo",
		Tags:     []string{"Genshin Impact", "Honkai Impact"},
	}

	storage := ipfsmock.NewMockIPFSStorage(t)
	storage.EXPECT().Version().Return("abcd", "abcd", nil).Once()

	ipfs := NewMockIPFSManager(storage)

	contract := fabricmock.NewMockContract()

	db := prepareTopicData(t)

	createTopic := invokeCreateTopic(logger, ipfs, db)

	t.Run("Creating Topic With Unmarshal Error", func(t *testing.T) {

		req := httptest.NewRequest(http.MethodPost, "/api/topic/invoke/CreateTopic", bytes.NewReader([]byte{1, 2, 3}))
		rec := httptest.NewRecorder()

		c := server.NewContext(req, rec)
		c = newMockSignedContext(c)

		err := createTopic(contract, c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("Creating Topic With Invalid Tags", func(t *testing.T) {

		payload.Tags = []string{"Genshin Impact", "Honkai Impact", "Invalid Tag"}

		req := httptest.NewRequest(http.MethodPost, "/api/topic/invoke/CreateTopic", newJsonRequest(&payload))
		req.Header.Add(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		c := server.NewContext(req, rec)
		c = newMockSignedContext(c)

		err := createTopic(contract, c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)

		payload.Tags = []string{"Genshin Impact", "Honkai Impact"}

	})

	t.Run("Creating Topic With IPFS Failure", func(t *testing.T) {

		storage.On("Add", mock.Anything).Return("", errors.New("hello world")).Once()

		req := httptest.NewRequest(http.MethodPost, "/api/topic/invoke/CreateTopic", newJsonRequest(&payload))
		req.Header.Add(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		c := server.NewContext(req, rec)
		c = newMockSignedContext(c)

		err := createTopic(contract, c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)

	})

	t.Run("Creating Topic With Base64DecodeError", func(t *testing.T) {

		payload.Images = []string{"base64Error&*"}
		storage.On("Add", mock.Anything).Return("base64", nil).Once()

		req := httptest.NewRequest(http.MethodPost, "/api/topic/invoke/CreateTopic", newJsonRequest(&payload))
		req.Header.Add(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		c := server.NewContext(req, rec)
		c = newMockSignedContext(c)

		err := createTopic(contract, c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)

	})

	t.Run("Creating Topic With IPFS Error on Uploading Images", func(t *testing.T) {

		payload.Images = []string{base64.StdEncoding.EncodeToString([]byte("base64Error&*"))}

		storage.On("Add", mock.Anything).Return("base64", nil).Once()
		storage.On("Add", mock.Anything).Return("", errors.New("Hello world")).Once()

		req := httptest.NewRequest(http.MethodPost, "/api/topic/invoke/CreateTopic", newJsonRequest(&payload))
		req.Header.Add(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		c := server.NewContext(req, rec)
		c = newMockSignedContext(c)

		err := createTopic(contract, c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})

	t.Run("Creating Topic With Chaincode Network Failure", func(t *testing.T) {

		payload.Images = []string{base64.StdEncoding.EncodeToString([]byte("base64Error&*"))}

		storage.On("Add", mock.Anything).Return("base64", nil)

		req := httptest.NewRequest(http.MethodPost, "/api/topic/invoke/CreateTopic", newJsonRequest(&payload))
		req.Header.Add(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		c := server.NewContext(req, rec)
		c = newMockSignedContext(c)

		contract.On("Submit", "CreateTopic", mock.Anything).Return([]byte(nil), errors.New("Hello world")).Once()
		err := createTopic(contract, c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)

	})

	t.Run("Creating Topic With Success", func(t *testing.T) {

		payload.Images = []string{base64.StdEncoding.EncodeToString([]byte("base64Error&*"))}

		storage.On("Add", mock.Anything).Return("base64", nil)

		req := httptest.NewRequest(http.MethodPost, "/api/topic/invoke/CreateTopic", newJsonRequest(&payload))
		req.Header.Add(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		c := server.NewContext(req, rec)
		c = newMockSignedContext(c)

		contract.On("Submit", "CreateTopic", mock.Anything).Return([]byte(nil), nil).Once()
		err := createTopic(contract, c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
	})
}

func TestCreateTopicCallback(t *testing.T) {

	topicBlock := TopicBlock{
		Title:    "this is a testing block",
		CID:      "abcd",
		Hash:     "abcd",
		Creator:  "0x123456789",
		Category: "Mihoyo",
		Tags:     []string{"Genshin Impact", "Honkai Impact"},
		Images:   []string{"abcd"},
	}

	storage := ipfsmock.NewMockIPFSStorage(t)
	storage.EXPECT().Version().Return("abcd", "abcd", nil).Once()

	ipfs := NewMockIPFSManager(storage)

	db := prepareTopicData(t)

	createTopic := createTopicCallback(logger, ipfs, db)

	t.Run("Create Topic Callback With IPFS failure", func(t *testing.T) {

		storage.EXPECT().Cat(topicBlock.CID).Return(io.ReadCloser(nil), errors.New("hello world")).Once()
		b, _ := json.Marshal(&topicBlock)

		err := createTopic(b)
		assert.Error(t, err)

	})

	reader := io.NopCloser(bytes.NewReader([]byte("document")))

	t.Run("Creating Topic Callback with user not found", func(t *testing.T) {

		storage.EXPECT().Cat(topicBlock.CID).Return(reader, nil)
		topicBlock.Creator = "unknown"
		b, _ := json.Marshal(&topicBlock)

		err := createTopic(b)
		assert.Error(t, err)
	})

	reader = io.NopCloser(bytes.NewReader([]byte("document")))

	t.Run("Creating Topic Callback with success", func(t *testing.T) {

		storage.EXPECT().Cat(topicBlock.CID).Return(reader, nil)
		topicBlock.Creator = "0x123456789"
		topicBlock.Category = "Mihoyo"
		topicBlock.Tags = []string{"Genshin Impact", "Honkai Impact"}

		b, _ := json.Marshal(&topicBlock)

		err := createTopic(b)
		assert.NoError(t, err)
	})

}

func TestInvokeDeleteTopic(t *testing.T) {
	type DeleteRequest struct {
		Hash string `json:"hash"`
	}

	payload := DeleteRequest{
		Hash: "topic1",
	}

	contract := fabricmock.NewMockContract()
	db := prepareTopicData(t)

	deleteTopic := invokeDeleteTopic(logger, db)

	t.Run("Deleting Topic With Unmarshal Error", func(t *testing.T) {

		req := httptest.NewRequest(http.MethodPost, "/api/topic/invoke/DeleteTopic", bytes.NewReader([]byte{1, 2, 3}))
		rec := httptest.NewRecorder()

		c := server.NewContext(req, rec)
		c = newMockSignedContext(c)

		err := deleteTopic(contract, c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("Deleting Topic With Hash Error", func(t *testing.T) {
		payload.Hash = "a111"
		req := httptest.NewRequest(http.MethodPost, "/api/topic/invoke/DeleteTopic", newJsonRequest(&payload))
		req.Header.Add(echo.HeaderContentType, echo.MIMEApplicationJSON)

		rec := httptest.NewRecorder()
		c := server.NewContext(req, rec)
		c = newMockSignedContext(c)

		err := deleteTopic(contract, c)

		assert.Error(t, err)
		payload.Hash = "topic1"
	})

	t.Run("Deleting Topic With Chaincode Network Failure", func(t *testing.T) {

		req := httptest.NewRequest(http.MethodPost, "/api/topic/invoke/DeleteTopic", newJsonRequest(&payload))
		req.Header.Add(echo.HeaderContentType, echo.MIMEApplicationJSON)

		rec := httptest.NewRecorder()
		c := server.NewContext(req, rec)
		c = newMockSignedContext(c)

		contract.On("Submit", "DeleteTopic", mock.Anything).Return([]byte(nil), errors.New("Hello world")).Once()
		err := deleteTopic(contract, c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)

	})

	t.Run("Deleting Topic With Success", func(t *testing.T) {

		req := httptest.NewRequest(http.MethodPost, "/api/topic/invoke/DeleteTopic", newJsonRequest(&payload))
		req.Header.Add(echo.HeaderContentType, echo.MIMEApplicationJSON)

		rec := httptest.NewRecorder()
		c := server.NewContext(req, rec)
		c = newMockSignedContext(c)

		contract.On("Submit", "DeleteTopic", mock.Anything).Return([]byte(nil), nil).Once()
		err := deleteTopic(contract, c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
	})
}

func TestDeleteTopicCallback(t *testing.T) {
	payload := DeleteBlock{
		Hash:    "topic1",
		Creator: "0x123456789",
	}

	db := prepareTopicData(t)
	deleteTopic := deleteTopicCallback(logger, db)

	t.Run("Deleting Topic Callback with hash not found", func(t *testing.T) {
		payload.Hash = "unknown"

		b, _ := json.Marshal(&payload)

		err := deleteTopic(b)
		assert.Error(t, err)
		payload.Hash = "topic1"
	})

	t.Run("Deleting Topic Callback with success", func(t *testing.T) {

		b, _ := json.Marshal(&payload)

		err := deleteTopic(b)
		assert.NoError(t, err)
	})
}

func TestInvokeUpdateTopic(t *testing.T) {
	type UpdateTopicRequest struct {
		Hash    string   `json:"hash"`
		Content string   `json:"content"`
		Images  []string `json:"assets"`
		Type    string   `json:"type"`
	}

	payload := UpdateTopicRequest{
		Hash:    "topic1",
		Content: "Hello world",
		Images:  []string{},
		Type:    "topic",
	}

	storage := ipfsmock.NewMockIPFSStorage(t)
	storage.EXPECT().Version().Return("abcd", "abcd", nil).Once()

	ipfs := NewMockIPFSManager(storage)

	contract := fabricmock.NewMockContract()

	db := prepareTopicData(t)

	updateTopic := invokeUpdateTopic(logger, ipfs, db)
	t.Run("Updating Topic With Unmarshal Error", func(t *testing.T) {

		req := httptest.NewRequest(http.MethodPost, "/api/topic/invoke/UpdateTopic", bytes.NewReader([]byte{1, 2, 3}))
		rec := httptest.NewRecorder()

		c := server.NewContext(req, rec)
		c = newMockSignedContext(c)

		err := updateTopic(contract, c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("Updating Topic With Hash Error", func(t *testing.T) {
		payload.Hash = "a111"
		req := httptest.NewRequest(http.MethodPost, "/api/topic/invoke/UpdateTopic", newJsonRequest(&payload))
		req.Header.Add(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		c := server.NewContext(req, rec)
		c = newMockSignedContext(c)

		err := updateTopic(contract, c)

		assert.Error(t, err)
		payload.Hash = "topic1"
	})

	t.Run("Updating Topic With IPFS Failure", func(t *testing.T) {

		storage.On("Add", mock.Anything).Return("", errors.New("hello world")).Once()

		req := httptest.NewRequest(http.MethodPost, "/api/topic/invoke/UpdateTopic", newJsonRequest(&payload))
		req.Header.Add(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		c := server.NewContext(req, rec)
		c = newMockSignedContext(c)

		err := updateTopic(contract, c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)

	})

	t.Run("Updating Topic With Base64DecodeError", func(t *testing.T) {

		payload.Images = []string{"base64Error&*"}
		storage.On("Add", mock.Anything).Return("base64", nil).Once()

		req := httptest.NewRequest(http.MethodPost, "/api/topic/invoke/UpdateTopic", newJsonRequest(&payload))
		req.Header.Add(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		c := server.NewContext(req, rec)
		c = newMockSignedContext(c)

		err := updateTopic(contract, c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)

	})

	t.Run("Updating Topic With IPFS Error on Uploading Images", func(t *testing.T) {

		payload.Images = []string{base64.StdEncoding.EncodeToString([]byte("base64Error&*"))}

		storage.On("Add", mock.Anything).Return("base64", nil).Once()
		storage.On("Add", mock.Anything).Return("", errors.New("Hello world")).Once()

		req := httptest.NewRequest(http.MethodPost, "/api/topic/invoke/UpdateTopic", newJsonRequest(&payload))
		req.Header.Add(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		c := server.NewContext(req, rec)
		c = newMockSignedContext(c)

		err := updateTopic(contract, c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})

	t.Run("Updating Topic With Chaincode Network Failure", func(t *testing.T) {

		payload.Images = []string{base64.StdEncoding.EncodeToString([]byte("base64Error&*"))}

		storage.On("Add", mock.Anything).Return("base64", nil)

		req := httptest.NewRequest(http.MethodPost, "/api/topic/invoke/UpdateTopic", newJsonRequest(&payload))
		req.Header.Add(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		c := server.NewContext(req, rec)
		c = newMockSignedContext(c)

		contract.On("Submit", "UpdateTopic", mock.Anything).Return([]byte(nil), errors.New("Hello world")).Once()
		err := updateTopic(contract, c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)

	})

	t.Run("Updating Topic With Success", func(t *testing.T) {

		payload.Images = []string{base64.StdEncoding.EncodeToString([]byte("base64Error&*"))}

		storage.On("Add", mock.Anything).Return("base64", nil)

		req := httptest.NewRequest(http.MethodPost, "/api/topic/invoke/UpdateTopic", newJsonRequest(&payload))
		req.Header.Add(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		c := server.NewContext(req, rec)
		c = newMockSignedContext(c)

		contract.On("Submit", "UpdateTopic", mock.Anything).Return([]byte(nil), nil).Once()
		err := updateTopic(contract, c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
	})

}

func TestUpdateTopicCallback(t *testing.T) {

	topicBlock := TopicBlock{
		CID:     "defg",
		Hash:    "topic1",
		Creator: "0x123456789",
		Images:  []string{"abcd"},
	}

	storage := ipfsmock.NewMockIPFSStorage(t)
	storage.EXPECT().Version().Return("abcd", "abcd", nil).Once()

	ipfs := NewMockIPFSManager(storage)

	db := prepareTopicData(t)

	updateTopic := updateTopicCallback(logger, ipfs, db)

	t.Run("Update Topic Callback With IPFS failure", func(t *testing.T) {

		storage.EXPECT().Cat(topicBlock.CID).Return(io.ReadCloser(nil), errors.New("hello world")).Once()
		b, _ := json.Marshal(&topicBlock)

		err := updateTopic(b)
		assert.Error(t, err)

	})

	reader := io.NopCloser(bytes.NewReader([]byte("document")))

	t.Run("Creating Topic Callback with success", func(t *testing.T) {

		storage.EXPECT().Cat(topicBlock.CID).Return(reader, nil)

		b, _ := json.Marshal(&topicBlock)

		err := updateTopic(b)
		assert.NoError(t, err)
	})

	t.Run("Creating Topic Callback with hash not found", func(t *testing.T) {
		topicBlock.Hash = "unknown"

		b, _ := json.Marshal(&topicBlock)

		err := updateTopic(b)
		assert.Error(t, err)
	})

}

func TestInvokeUpvoteTopic(t *testing.T) {
	type UpvoteRequest struct {
		Hash string `json:"hash"`
		Type string `json:"type"`
	}

	payload := UpvoteRequest{
		Hash: "topic1",
		Type: "Topic",
	}

	contract := fabricmock.NewMockContract()

	db := prepareTopicData(t)

	upvoteTopic := invokeUpvoteTopic(logger, db)

	t.Run("Upvoting Topic With Unmarshal Error", func(t *testing.T) {

		req := httptest.NewRequest(http.MethodPost, "/api/topic/invoke/upvote", bytes.NewReader([]byte{1, 2, 3}))
		rec := httptest.NewRecorder()

		c := server.NewContext(req, rec)
		c = newMockSignedContext(c)

		err := upvoteTopic(contract, c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("Upvoting Topic With Hash Error", func(t *testing.T) {
		payload.Hash = "a111"
		req := httptest.NewRequest(http.MethodPost, "/api/topic/invoke/upvote", newJsonRequest(&payload))
		req.Header.Add(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		c := server.NewContext(req, rec)
		c = newMockSignedContext(c)

		err := upvoteTopic(contract, c)

		assert.Error(t, err)
		payload.Hash = "topic1"
	})

	t.Run("Upvoting Topic With Chaincode Network Failure", func(t *testing.T) {

		req := httptest.NewRequest(http.MethodPost, "/api/topic/invoke/upvote", newJsonRequest(&payload))
		req.Header.Add(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		c := server.NewContext(req, rec)
		c = newMockSignedContext(c)

		contract.On("Submit", "UpvoteTopic", mock.Anything).Return([]byte(nil), errors.New("Hello world")).Once()
		err := upvoteTopic(contract, c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)

	})

	t.Run("Upvoting Topic With Success", func(t *testing.T) {

		req := httptest.NewRequest(http.MethodPost, "/api/topic/invoke/upvote", newJsonRequest(&payload))
		req.Header.Add(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		c := server.NewContext(req, rec)
		c = newMockSignedContext(c)

		contract.On("Submit", "UpvoteTopic", mock.Anything).Return([]byte(nil), nil).Once()
		err := upvoteTopic(contract, c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
	})
}

func TestUpvoteTopicCallback(t *testing.T) {

	upvoteBlock := UpvoteBlock{
		Hash:    "topic1",
		Creator: "0x123456789",
	}

	db := prepareTopicData(t)

	upvoteTopic := upvoteTopicCallback(logger, db)

	t.Run("Upvoting Topic Callback with hash not found", func(t *testing.T) {
		upvoteBlock.Hash = "unknown"

		b, _ := json.Marshal(&upvoteBlock)

		err := upvoteTopic(b)
		assert.Error(t, err)
		upvoteBlock.Hash = "topic1"
	})

	t.Run("Upvoting Topic Callback with success", func(t *testing.T) {

		b, _ := json.Marshal(&upvoteBlock)

		err := upvoteTopic(b)
		assert.NoError(t, err)
	})

}

func TestInvokeDownvoteTopic(t *testing.T) {
	type DownvoteRequest struct {
		Hash string `json:"hash"`
		Type string `json:"type"`
	}

	payload := DownvoteRequest{
		Hash: "topic1",
		Type: "Topic",
	}

	contract := fabricmock.NewMockContract()

	db := prepareTopicData(t)

	downvoteTopic := invokeDownvoteTopic(logger, db)

	t.Run("Downvoting Topic With Unmarshal Error", func(t *testing.T) {

		req := httptest.NewRequest(http.MethodPost, "/api/topic/invoke/downvote", bytes.NewReader([]byte{1, 2, 3}))
		rec := httptest.NewRecorder()

		c := server.NewContext(req, rec)
		c = newMockSignedContext(c)

		err := downvoteTopic(contract, c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("Downvoting Topic With Hash Error", func(t *testing.T) {
		payload.Hash = "a111"
		req := httptest.NewRequest(http.MethodPost, "/api/topic/invoke/downvote", newJsonRequest(&payload))
		req.Header.Add(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		c := server.NewContext(req, rec)
		c = newMockSignedContext(c)

		err := downvoteTopic(contract, c)

		assert.Error(t, err)
		payload.Hash = "topic1"
	})

	t.Run("Downvoting Topic With Chaincode Network Failure", func(t *testing.T) {

		req := httptest.NewRequest(http.MethodPost, "/api/topic/invoke/downvote", newJsonRequest(&payload))
		req.Header.Add(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		c := server.NewContext(req, rec)
		c = newMockSignedContext(c)

		contract.On("Submit", "DownvoteTopic", mock.Anything).Return([]byte(nil), errors.New("Hello world")).Once()
		err := downvoteTopic(contract, c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)

	})

	t.Run("Downvoting Topic With Success", func(t *testing.T) {

		req := httptest.NewRequest(http.MethodPost, "/api/topic/invoke/downvote", newJsonRequest(&payload))
		req.Header.Add(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		c := server.NewContext(req, rec)
		c = newMockSignedContext(c)

		contract.On("Submit", "DownvoteTopic", mock.Anything).Return([]byte(nil), nil).Once()
		err := downvoteTopic(contract, c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
	})
}

func TestDownvoteTopicCallback(t *testing.T) {

	downvoteBlock := DownvoteBlock{
		Hash:    "topic1",
		Creator: "0x123456789",
	}

	db := prepareTopicData(t)

	downvoteTopic := downvoteTopicCallback(logger, db)

	t.Run("Downvoting Topic Callback with hash not found", func(t *testing.T) {
		downvoteBlock.Hash = "unknown"

		b, _ := json.Marshal(&downvoteBlock)

		err := downvoteTopic(b)
		assert.Error(t, err)
		downvoteBlock.Hash = "topic1"
	})

	t.Run("Downvoting Topic Callback with success", func(t *testing.T) {

		b, _ := json.Marshal(&downvoteBlock)

		err := downvoteTopic(b)
		assert.NoError(t, err)
	})

}

func TestQueryCategories(t *testing.T) {
	t.Run("Querying Categories With Success", func(t *testing.T) {

		db := newSqliteDB()

		query := queryCategories(logger, db)
		results := []CategoryGroup{}

		req := httptest.NewRequest(http.MethodGet, "/api/topic/query/categories", newJsonRequest(&results))
		req.Header.Add(echo.HeaderContentType, echo.MIMEApplicationJSON)

		rec := httptest.NewRecorder()
		c := server.NewContext(req, rec)

		var _ = query(c)
		assert.Equal(t, http.StatusOK, rec.Code)
	})
}

func TestQueryTags(t *testing.T) {
	t.Run("Querying Tags With Success", func(t *testing.T) {

		db := newSqliteDB()

		query := queryTags(logger, db)
		results := []Topic{}

		req := httptest.NewRequest(http.MethodGet, "/api/topic/query/tags", newJsonRequest(&results))
		req.Header.Add(echo.HeaderContentType, echo.MIMEApplicationJSON)

		rec := httptest.NewRecorder()
		c := server.NewContext(req, rec)

		var _ = query(c)
		assert.Equal(t, http.StatusOK, rec.Code)
	})
}

func TestQueryTopicGet(t *testing.T) {

	db := prepareTopicData(t)
	query := queryTopicGet(logger, db)
	t.Run("Downvoting Topic With Unmarshal Error", func(t *testing.T) {

		req := httptest.NewRequest(http.MethodPost, "/api/topic/query/get", bytes.NewReader([]byte{1, 2, 3}))
		rec := httptest.NewRecorder()
		c := server.NewContext(req, rec)

		var _ = query(c)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("Querying Topic With Success", func(t *testing.T) {

		type TopicGetRequest struct {
			Hash string `json:"hash"`
		}

		payload := TopicGetRequest{
			Hash: "topic1",
		}
		req := httptest.NewRequest(http.MethodPost, "/api/topic/query/get", newJsonRequest(&payload))
		req.Header.Add(echo.HeaderContentType, echo.MIMEApplicationJSON)

		rec := httptest.NewRecorder()
		c := server.NewContext(req, rec)

		var _ = query(c)
		assert.Equal(t, http.StatusOK, rec.Code)
	})
}

func TestNewTopicChaincodeMiddleware(t *testing.T) {

	network := mocks.NewMockNetwork(t)

	storage := ipfsmock.NewMockIPFSStorage(t)
	storage.EXPECT().Version().Return("abcd", "abcd", nil).Once()

	ipfs := NewMockIPFSManager(storage)

	db := newSqliteDB()
	network.EXPECT().GetContract("topic").Return(&client.Contract{}).Once()
	var _ = NewTopicChaincodeMiddleware(logger, network, ipfs, db)

}
