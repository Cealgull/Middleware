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

	user2 := &User{
		Username: "User",
		Wallet:   "0x1000000",
		Avatar:   "null",
	}

	tags := []*Tag{
		{Name: "Genshin Impact", Creator: user, Description: "Genshin Impact"},
		{Name: "Honkai Impact", Creator: user, Description: "Honkai Impact"},
	}

	topics := []*Topic{
		{
			Hash:          "topic1",
			Title:         "This is a testing topic",
			CreatorWallet: "0x123456789",
			Creator:       user,
			Content:       "Hello world",
		},
		{
			Hash:          "topic2",
			Title:         "This is a testing topic",
			CreatorWallet: "0x123456789",
			Creator:       user,
			Content:       "Hello world",
			Upvotes: []*Upvote{
				{
					CreatorWallet: user.Wallet,
				},
			},
		},
		{
			Hash:          "topic3",
			Title:         "This is a testing topic",
			CreatorWallet: "0x123456789",
			Creator:       user2,
			Content:       "Hello world",
			Downvotes: []*Downvote{
				{
					CreatorWallet: user.Wallet,
				},
			},
		},
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
	assert.NoError(t, db.Create(&topics).Error)

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

	t.Run("Creating Topic With Invalid Category", func(t *testing.T) {

		payload.Category = "Invalid Category"

		req := httptest.NewRequest(http.MethodPost, "/api/topic/invoke/CreateTopic", newJsonRequest(&payload))
		req.Header.Add(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		c := server.NewContext(req, rec)
		c = newMockSignedContext(c)

		err := createTopic(contract, c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)

		payload.Category = "Mihoyo"
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

	t.Run("Creating Topic With Chaincode Network Failure", func(t *testing.T) {

		payload.Images = []string{base64.StdEncoding.EncodeToString([]byte("base64Error&*"))}

		storage.On("Add", mock.Anything).Return("base64", nil)

		req := httptest.NewRequest(http.MethodPost, "/api/topic/invoke/CreateTopic", newJsonRequest(&payload))
		req.Header.Add(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		c := server.NewContext(req, rec)
		c = newMockSignedContext(c)

		contract.On("SubmitAsync", "CreateTopic", mock.Anything).Return([]byte(nil), (*client.Commit)(nil), errors.New("Hello world")).Once()
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

		contract.On("SubmitAsync", "CreateTopic", mock.Anything).Return([]byte(nil), (*client.Commit)(nil), nil).Once()
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

		contract.On("SubmitAsync", "DeleteTopic", mock.Anything).Return([]byte(nil), (*client.Commit)(nil), errors.New("Hello world")).Once()
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

		contract.On("SubmitAsync", "DeleteTopic", mock.Anything).Return([]byte(nil), (*client.Commit)(nil), nil).Once()
		err := deleteTopic(contract, c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("Deleting Other's Topic", func(t *testing.T) {

		payload.Hash = "topic3"

		req := httptest.NewRequest(http.MethodPost, "/api/topic/invoke/DeleteTopic", newJsonRequest(&payload))
		req.Header.Add(echo.HeaderContentType, echo.MIMEApplicationJSON)

		rec := httptest.NewRecorder()
		c := server.NewContext(req, rec)
		c = newMockSignedContext(c)

		contract.On("SubmitAsync", "DeleteTopic", mock.Anything).Return([]byte(nil), (*client.Commit)(nil), nil).Once()
		err := deleteTopic(contract, c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)

		payload.Hash = "topic1"
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
		Hash     string   `json:"hash"`
		Content  string   `json:"content"`
		Images   []string `json:"images"`
		Title    string   `json:"title"`
		Category string   `json:"category"`
		Tags     []string `json:"tags"`
	}

	payload := UpdateTopicRequest{
		Hash:     "topic1",
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

	t.Run("Updating Topic With Invalid Tags", func(t *testing.T) {
		payload.Tags = []string{"Genshin Impact", "Honkai Impact", "Invalid Tag"}
		req := httptest.NewRequest(http.MethodPost, "/api/topic/invoke/UpdateTopic", newJsonRequest(&payload))
		req.Header.Add(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		c := server.NewContext(req, rec)
		c = newMockSignedContext(c)

		err := updateTopic(contract, c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)

		payload.Tags = []string{"Genshin Impact", "Honkai Impact"}
	})

	t.Run("Updating Topic With Invalid Category", func(t *testing.T) {
		payload.Category = "Invalid Category"
		req := httptest.NewRequest(http.MethodPost, "/api/topic/invoke/UpdateTopic", newJsonRequest(&payload))
		req.Header.Add(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		c := server.NewContext(req, rec)
		c = newMockSignedContext(c)

		err := updateTopic(contract, c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		payload.Category = "Mihoyo"
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

	t.Run("Updating Topic With Chaincode Network Failure", func(t *testing.T) {

		payload.Images = []string{base64.StdEncoding.EncodeToString([]byte("base64Error&*"))}

		storage.On("Add", mock.Anything).Return("base64", nil)

		req := httptest.NewRequest(http.MethodPost, "/api/topic/invoke/UpdateTopic", newJsonRequest(&payload))
		req.Header.Add(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		c := server.NewContext(req, rec)
		c = newMockSignedContext(c)

		contract.On("SubmitAsync", "UpdateTopic", mock.Anything).Return([]byte(nil), (*client.Commit)(nil), errors.New("Hello world")).Once()
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

		contract.On("SubmitAsync", "UpdateTopic", mock.Anything).Return([]byte(nil), (*client.Commit)(nil), nil).Once()
		err := updateTopic(contract, c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
	})

}

func TestUpdateTopicCallback(t *testing.T) {

	topicBlock := TopicBlock{
		CID:      "defg",
		Hash:     "topic1",
		Creator:  "0x123456789",
		Images:   []string{"abcd"},
		Category: "Mihoyo",
		Tags:     []string{"Genshin Impact", "Honkai Impact"},
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

	t.Run("Update Topic Callback with success", func(t *testing.T) {

		storage.EXPECT().Cat(topicBlock.CID).Return(reader, nil)

		b, _ := json.Marshal(&topicBlock)

		err := updateTopic(b)
		assert.NoError(t, err)
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

		contract.On("SubmitAsync", "UpvoteTopic", mock.Anything).Return([]byte(nil), (*client.Commit)(nil), errors.New("Hello world")).Once()
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

		contract.On("SubmitAsync", "UpvoteTopic", mock.Anything).Return([]byte(nil), (*client.Commit)(nil), nil).Once()
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

	t.Run("Upvoting Topic Callback with Upvoted", func(t *testing.T) {

		upvoteBlock.Hash = "topic2"
		b, _ := json.Marshal(&upvoteBlock)

		err := upvoteTopic(b)
		assert.NoError(t, err)

		// Verify that the upvote has been removed from the topic
		topic := Topic{}
		assert.NoError(t, db.Preload("Upvotes").Where("hash = ?", upvoteBlock.Hash).First(&topic).Error)
		assert.Empty(t, topic.Upvotes) // Ensure Upvotes is empty after removal

		upvoteBlock.Hash = "topic1"
	})

	t.Run("Upvoting Topic Callback with Downvoted", func(t *testing.T) {

		upvoteBlock.Hash = "topic3"
		b, _ := json.Marshal(&upvoteBlock)

		err := upvoteTopic(b)
		assert.NoError(t, err)

		// Verify that the downvote has been removed from the topic
		topic := Topic{}
		assert.NoError(t, db.Preload("Downvotes").Where("hash = ?", upvoteBlock.Hash).First(&topic).Error)
		assert.Empty(t, topic.Downvotes) // Ensure Downvotes is empty after removal

		upvoteBlock.Hash = "topic1"
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

		contract.On("SubmitAsync", "DownvoteTopic", mock.Anything).Return([]byte(nil), (*client.Commit)(nil), errors.New("Hello world")).Once()
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

		contract.On("SubmitAsync", "DownvoteTopic", mock.Anything).Return([]byte(nil), (*client.Commit)(nil), nil).Once()
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

	t.Run("Downvoting Topic Callback with Upvoted", func(t *testing.T) {

		downvoteBlock.Hash = "topic2"
		b, _ := json.Marshal(&downvoteBlock)

		err := downvoteTopic(b)
		assert.NoError(t, err)
		downvoteBlock.Hash = "topic1"
	})

	t.Run("Downvoting Topic Callback with Downvoted", func(t *testing.T) {

		downvoteBlock.Hash = "topic3"
		b, _ := json.Marshal(&downvoteBlock)

		err := downvoteTopic(b)
		assert.NoError(t, err)
		downvoteBlock.Hash = "topic1"
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

	t.Run("Querying Topic With Failure", func(t *testing.T) {

		type TopicGetRequest struct {
			Hash string `json:"hash"`
		}

		payload := TopicGetRequest{
			Hash: "topic5",
		}
		req := httptest.NewRequest(http.MethodPost, "/api/topic/query/get", newJsonRequest(&payload))
		req.Header.Add(echo.HeaderContentType, echo.MIMEApplicationJSON)

		rec := httptest.NewRecorder()
		c := server.NewContext(req, rec)

		var _ = query(c)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}

func TestQueryTopicsList(t *testing.T) {
	type QueryRequest struct {
		PageOrdinal int      `json:"pageOrdinal"`
		PageSize    int      `json:"pageSize"`
		Category    string   `json:"category"`
		Creator     string   `json:"creator"`
		Tags        []string `json:"tags"`
	}

	payload := QueryRequest{
		PageOrdinal: 1,
		PageSize:    10,
		Category:    "Mihoyo",
		Creator:     "0x123456789",
		Tags:        []string{"Genshin Impact", "Honkai Impact"},
	}

	t.Run("Query Topic List With Unmarshal Error", func(t *testing.T) {

		db := newSqliteDB()

		query := queryTopicsList(logger, db)

		req := httptest.NewRequest(http.MethodPost, "/api/topic/query/list", bytes.NewReader([]byte{1, 2, 3}))
		rec := httptest.NewRecorder()
		c := server.NewContext(req, rec)

		err := query(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("Querying List With Illegal PageOrdinal", func(t *testing.T) {

		db := newSqliteDB()

		query := queryTopicsList(logger, db)

		payload.PageOrdinal = -1
		req := httptest.NewRequest(http.MethodPost, "/api/topic/query/list", newJsonRequest(&payload))
		req.Header.Add(echo.HeaderContentType, echo.MIMEApplicationJSON)

		rec := httptest.NewRecorder()
		c := server.NewContext(req, rec)

		err := query(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		payload.PageOrdinal = 1
	})

	t.Run("Querying List With Success", func(t *testing.T) {

		db := newSqliteDB()

		query := queryTopicsList(logger, db)

		req := httptest.NewRequest(http.MethodPost, "/api/topic/query/list", newJsonRequest(&payload))
		req.Header.Add(echo.HeaderContentType, echo.MIMEApplicationJSON)

		rec := httptest.NewRecorder()
		c := server.NewContext(req, rec)

		err := query(c)
		assert.NoError(t, err)
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
