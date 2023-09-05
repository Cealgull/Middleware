package chaincodes

import (
	"bytes"
	"encoding/json"
	"errors"
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

func prepareTagData(t *testing.T) *gorm.DB {

	user := &User{
		Username: "Admin",
		Wallet:   "0x123456789",
		Avatar:   "null",
	}

	tags := []*Tag{
		{Name: "Genshin Impact", Creator: user, Description: "Genshin Impact"},
		{Name: "Honkai Impact", Creator: user, Description: "Honkai Impact"},
	}

	db := newSqliteDB()

	assert.NoError(t, db.Create(&user).Error)
	assert.NoError(t, db.Create(&tags).Error)
	assert.NoError(t, db.Create(&CategoryGroup{
		Name:  "Games",
		Color: 123456,
		Categories: []*Category{{
			Name:  "Mihoyo",
			Color: 123456,
		},
		},
	}).Error)

	return db
}

func TestInvokeCreateTag(t *testing.T) {

	type TagRequest struct {
		Name        string `json:"name"`
		CreatorID   uint   `json:"creatorID"`
		Description string `json:"description"`
	}

	payload := TagRequest{
		Name:        "testing tag",
		CreatorID:   1,
		Description: "This is a testing tag.",
	}

	storage := ipfsmock.NewMockIPFSStorage(t)
	storage.EXPECT().Version().Return("abcd", "abcd", nil).Once()

	ipfs := NewMockIPFSManager(storage)

	contract := fabricmock.NewMockContract()

	db := prepareTagData(t)

	createTag := invokeCreateTag(logger, ipfs, db)

	t.Run("Creating Tag With Unmarshal Error", func(t *testing.T) {

		req := httptest.NewRequest(http.MethodPost, "/api/tag/invoke/CreateTag", bytes.NewReader([]byte{1, 2, 3}))
		rec := httptest.NewRecorder()

		c := server.NewContext(req, rec)
		c = newMockSignedContext(c)

		err := createTag(contract, c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("Creating Tag With Chaincode Network Failure", func(t *testing.T) {

		req := httptest.NewRequest(http.MethodPost, "/api/tag/invoke/CreateTag", newJsonRequest(&payload))
		req.Header.Add(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		c := server.NewContext(req, rec)
		c = newMockSignedContext(c)

		contract.On("Submit", "CreateTag", mock.Anything).Return([]byte(nil), errors.New("Hello world")).Once()
		err := createTag(contract, c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)

	})

	t.Run("Creating Tag With Success", func(t *testing.T) {

		req := httptest.NewRequest(http.MethodPost, "/api/tag/invoke/CreateTag", newJsonRequest(&payload))
		req.Header.Add(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		c := server.NewContext(req, rec)
		c = newMockSignedContext(c)

		contract.On("Submit", "CreateTag", mock.Anything).Return([]byte(nil), nil).Once()
		err := createTag(contract, c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
	})
}

func TestCreateTagCallback(t *testing.T) {

	tagBlock := TagBlock{
		Name:        "testing tag",
		CreatorID:   1,
		Description: "This is a testing tag.",
	}

	storage := ipfsmock.NewMockIPFSStorage(t)
	storage.EXPECT().Version().Return("abcd", "abcd", nil).Once()

	ipfs := NewMockIPFSManager(storage)

	db := prepareTagData(t)

	createTag := createTagCallback(logger, ipfs, db)

	// reader := io.NopCloser(bytes.NewReader([]byte("document")))

	t.Run("Creating Tag Callback with user not found", func(t *testing.T) {

		tagBlock.CreatorID = 100
		b, _ := json.Marshal(&tagBlock)

		err := createTag(b)
		assert.Error(t, err)
	})

	// reader = io.NopCloser(bytes.NewReader([]byte("document")))

	t.Run("Creating Tag Callback with success", func(t *testing.T) {

		tagBlock.CreatorID = 1

		b, _ := json.Marshal(&tagBlock)

		err := createTag(b)
		assert.NoError(t, err)
	})

}

func TestNewTagChaincodeMiddleware(t *testing.T) {

	network := mocks.NewMockNetwork(t)

	storage := ipfsmock.NewMockIPFSStorage(t)
	storage.EXPECT().Version().Return("abcd", "abcd", nil).Once()

	ipfs := NewMockIPFSManager(storage)

	db := newSqliteDB()
	network.EXPECT().GetContract("tag").Return(&client.Contract{}).Once()
	var _ = NewTagChaincodeMiddleware(logger, network, ipfs, db)

}

func prepareCategoryData(t *testing.T) *gorm.DB {

	user := &User{
		Username: "Admin",
		Wallet:   "0x123456789",
		Avatar:   "null",
	}

	db := newSqliteDB()

	assert.NoError(t, db.Create(&user).Error)
	assert.NoError(t, db.Create(&CategoryGroup{
		Name:  "Games",
		Color: 123456,
		Categories: []*Category{{
			Name:  "Mihoyo",
			Color: 123456,
		},
		},
	}).Error)

	return db
}

func TestInvokeCreateCategory(t *testing.T) {

	type CategoryRequest struct {
		CategoryGroupID uint   `json:"categoryGroupID"`
		Color           uint   `json:"color"`
		Name            string `json:"name"`
	}

	payload := CategoryRequest{
		CategoryGroupID: 1,
		Color:           1,
		Name:            "testing category",
	}

	storage := ipfsmock.NewMockIPFSStorage(t)
	storage.EXPECT().Version().Return("abcd", "abcd", nil).Once()

	ipfs := NewMockIPFSManager(storage)

	contract := fabricmock.NewMockContract()

	db := prepareCategoryData(t)

	createCategory := invokeCreateCategory(logger, ipfs, db)

	t.Run("Creating Category With Unmarshal Error", func(t *testing.T) {

		req := httptest.NewRequest(http.MethodPost, "/api/category/invoke/CreateCategory", bytes.NewReader([]byte{1, 2, 3}))
		rec := httptest.NewRecorder()

		c := server.NewContext(req, rec)
		c = newMockSignedContext(c)

		err := createCategory(contract, c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("Creating Category With Chaincode Network Failure", func(t *testing.T) {

		req := httptest.NewRequest(http.MethodPost, "/api/category/invoke/CreateCategory", newJsonRequest(&payload))
		req.Header.Add(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		c := server.NewContext(req, rec)
		c = newMockSignedContext(c)

		contract.On("Submit", "CreateCategory", mock.Anything).Return([]byte(nil), errors.New("Hello world")).Once()
		err := createCategory(contract, c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)

	})

	t.Run("Creating Category With Success", func(t *testing.T) {

		req := httptest.NewRequest(http.MethodPost, "/api/category/invoke/CreateCategory", newJsonRequest(&payload))
		req.Header.Add(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		c := server.NewContext(req, rec)
		c = newMockSignedContext(c)

		contract.On("Submit", "CreateCategory", mock.Anything).Return([]byte(nil), nil).Once()
		err := createCategory(contract, c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
	})
}

func TestCreateCategoryCallback(t *testing.T) {

	categoryBlock := CategoryBlock{
		CategoryGroupID: 1,
		Color:           1,
		Name:            "testing category",
	}

	storage := ipfsmock.NewMockIPFSStorage(t)
	storage.EXPECT().Version().Return("abcd", "abcd", nil).Once()

	ipfs := NewMockIPFSManager(storage)

	db := prepareCategoryData(t)

	createCategory := createCategoryCallback(logger, ipfs, db)

	// reader := io.NopCloser(bytes.NewReader([]byte("document")))

	t.Run("Creating Category Callback with categoryGroup not found", func(t *testing.T) {

		categoryBlock.CategoryGroupID = 100
		b, _ := json.Marshal(&categoryBlock)

		err := createCategory(b)
		assert.Error(t, err)
	})

	// reader = io.NopCloser(bytes.NewReader([]byte("document")))

	t.Run("Creating Category Callback with success", func(t *testing.T) {

		categoryBlock.CategoryGroupID = 1

		b, _ := json.Marshal(&categoryBlock)

		err := createCategory(b)
		assert.NoError(t, err)
	})

}

func TestNewCategoryChaincodeMiddleware(t *testing.T) {

	network := mocks.NewMockNetwork(t)

	storage := ipfsmock.NewMockIPFSStorage(t)
	storage.EXPECT().Version().Return("abcd", "abcd", nil).Once()

	ipfs := NewMockIPFSManager(storage)

	db := newSqliteDB()
	network.EXPECT().GetContract("category").Return(&client.Contract{}).Once()
	var _ = NewCategoryChaincodeMiddleware(logger, network, ipfs, db)

}

func prepareCategoryGroupData(t *testing.T) *gorm.DB {

	user := &User{
		Username: "Admin",
		Wallet:   "0x123456789",
		Avatar:   "null",
	}

	db := newSqliteDB()

	assert.NoError(t, db.Create(&user).Error)
	assert.NoError(t, db.Create(&CategoryGroup{
		Name:  "Games",
		Color: 123456,
		Categories: []*Category{{
			Name:  "Mihoyo",
			Color: 123456,
		},
		},
	}).Error)

	return db
}

func TestInvokeCreateCategoryGroup(t *testing.T) {

	type CategoryGroupRequest struct {
		CategoryGroupGroupID uint   `json:"categoryGroupGroupID"`
		Color                uint   `json:"color"`
		Name                 string `json:"name"`
	}

	payload := CategoryGroupRequest{
		CategoryGroupGroupID: 1,
		Color:                1,
		Name:                 "testing categoryGroup",
	}

	storage := ipfsmock.NewMockIPFSStorage(t)
	storage.EXPECT().Version().Return("abcd", "abcd", nil).Once()

	ipfs := NewMockIPFSManager(storage)

	contract := fabricmock.NewMockContract()

	db := prepareCategoryGroupData(t)

	createCategoryGroup := invokeCreateCategoryGroup(logger, ipfs, db)

	t.Run("Creating CategoryGroup With Unmarshal Error", func(t *testing.T) {

		req := httptest.NewRequest(http.MethodPost, "/api/categoryGroup/invoke/CreateCategoryGroup", bytes.NewReader([]byte{1, 2, 3}))
		rec := httptest.NewRecorder()

		c := server.NewContext(req, rec)
		c = newMockSignedContext(c)

		err := createCategoryGroup(contract, c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("Creating CategoryGroup With Chaincode Network Failure", func(t *testing.T) {

		req := httptest.NewRequest(http.MethodPost, "/api/categoryGroup/invoke/CreateCategoryGroup", newJsonRequest(&payload))
		req.Header.Add(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		c := server.NewContext(req, rec)
		c = newMockSignedContext(c)

		contract.On("Submit", "CreateCategoryGroup", mock.Anything).Return([]byte(nil), errors.New("Hello world")).Once()
		err := createCategoryGroup(contract, c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)

	})

	t.Run("Creating CategoryGroup With Success", func(t *testing.T) {

		req := httptest.NewRequest(http.MethodPost, "/api/categoryGroup/invoke/CreateCategoryGroup", newJsonRequest(&payload))
		req.Header.Add(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		c := server.NewContext(req, rec)
		c = newMockSignedContext(c)

		contract.On("Submit", "CreateCategoryGroup", mock.Anything).Return([]byte(nil), nil).Once()
		err := createCategoryGroup(contract, c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
	})
}

func TestCreateCategoryGroupCallback(t *testing.T) {

	categoryGroupBlock := CategoryGroupBlock{
		Color: 1,
		Name:  "testing categoryGroup",
	}

	storage := ipfsmock.NewMockIPFSStorage(t)
	storage.EXPECT().Version().Return("abcd", "abcd", nil).Once()

	ipfs := NewMockIPFSManager(storage)

	db := prepareCategoryGroupData(t)

	createCategoryGroup := createCategoryGroupCallback(logger, ipfs, db)

	// reader := io.NopCloser(bytes.NewReader([]byte("document")))

	// t.Run("Creating CategoryGroup Callback with categoryGroupGroup not found", func(t *testing.T) {

	// 	categoryGroupBlock.CategoryGroupGroupID = 100
	// 	b, _ := json.Marshal(&categoryGroupBlock)

	// 	err := createCategoryGroup(b)
	// 	assert.Error(t, err)
	// })

	// reader = io.NopCloser(bytes.NewReader([]byte("document")))

	t.Run("Creating CategoryGroup Callback with success", func(t *testing.T) {
		b, _ := json.Marshal(&categoryGroupBlock)

		err := createCategoryGroup(b)
		assert.NoError(t, err)
	})

}

func TestNewCategoryGroupChaincodeMiddleware(t *testing.T) {

	network := mocks.NewMockNetwork(t)

	storage := ipfsmock.NewMockIPFSStorage(t)
	storage.EXPECT().Version().Return("abcd", "abcd", nil).Once()

	ipfs := NewMockIPFSManager(storage)

	db := newSqliteDB()
	network.EXPECT().GetContract("categoryGroup").Return(&client.Contract{}).Once()
	var _ = NewCategoryGroupChaincodeMiddleware(logger, network, ipfs, db)

}
