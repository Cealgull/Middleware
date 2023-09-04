package chaincodes

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Cealgull/Middleware/internal/fabric/common/mocks"
	. "github.com/Cealgull/Middleware/internal/models"
	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func newJsonRequest(payload interface{}) io.Reader {
	b, _ := json.Marshal(payload)
	return bytes.NewReader(b)
}

var mockInvokeResult []byte = []byte{1, 2, 3, 4}

func TestInvokeCreateUser(t *testing.T) {

	contract := mocks.NewMockContract()

	i := invokeCreateUser(logger)

	t.Run("Invoke Creating User Normally", func(t *testing.T) {

		req := httptest.NewRequest(http.MethodPost, "/api/user/invoke/create", nil)
		rec := httptest.NewRecorder()
		c := server.NewContext(req, rec)
		c = newMockSignedContext(c)

		var _ = contract.On("Submit", "CreateUser", mock.Anything).Return(mockInvokeResult, nil).Once()
		err := i(contract, c)
		assert.True(t, contract.AssertExpectations(t))

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("Invoke Creating User Error", func(t *testing.T) {

		req := httptest.NewRequest(http.MethodPost, "/api/user/invoke/create", nil)
		rec := httptest.NewRecorder()
		c := server.NewContext(req, rec)
		c = newMockSignedContext(c)

		var _ = contract.On("Submit", "CreateUser", mock.Anything).Return([]byte(nil), errors.New("Submit Failure")).Once()
		err := i(contract, c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)

	})
}

func TestInvokeUpdateUser(t *testing.T) {

	type ProfileChanged struct {
		Username  string `json:"username"`
		Wallet    string `json:"wallet"`
		Avatar    string `json:"avatar"`
		Signature string `json:"signature"`

		ActiveRole  uint `json:"activeRole"`
		ActiveBadge uint `json:"activeBadge"`
	}

	contract := mocks.NewMockContract()

	db := newSqliteDB()
	u := invokeUpdateUser(logger, db)

	profileChanged := ProfileChanged{
		Username:    "Alice",
		Wallet:      "0x123456789",
		Avatar:      "null",
		Signature:   "null",
		ActiveRole:  1,
		ActiveBadge: 1,
	}

	t.Run("Invoke Updating User With No Content Header", func(t *testing.T) {

		req := httptest.NewRequest(http.MethodPost, "/api/user/invoke/update", newJsonRequest(&profileChanged))
		rec := httptest.NewRecorder()

		c := server.NewContext(req, rec)
		c = newMockSignedContext(c)

		err := u(contract, c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("Invoke Updating User With No Content Type", func(t *testing.T) {

		req := httptest.NewRequest(http.MethodPost, "/api/user/invoke/update", newJsonRequest(&profileChanged))
		rec := httptest.NewRecorder()

		c := server.NewContext(req, rec)
		c = newMockSignedContext(c)

		err := u(contract, c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)

	})

	t.Run("Invoke Updating User With Success", func(t *testing.T) {

		req := httptest.NewRequest(http.MethodPost, "/api/user/invoke/update", newJsonRequest(&profileChanged))
		req.Header.Add(echo.HeaderContentType, echo.MIMEApplicationJSON)

		rec := httptest.NewRecorder()

		c := server.NewContext(req, rec)
		c = newMockSignedContext(c)
		contract.On("Submit", "UpdateUser", mock.Anything).Return(mockInvokeResult, nil).Once()
		err := u(contract, c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

	})

	t.Run("Invoke Updating User With Failure", func(t *testing.T) {

		req := httptest.NewRequest(http.MethodPost, "/api/user/invoke/update", newJsonRequest(&profileChanged))
		req.Header.Add(echo.HeaderContentType, echo.MIMEApplicationJSON)

		rec := httptest.NewRecorder()

		c := server.NewContext(req, rec)
		c = newMockSignedContext(c)
		contract.On("Submit", "UpdateUser", mock.Anything).Return([]byte(nil), errors.New("Invoke Failure")).Once()
		err := u(contract, c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)

	})
}

func TestAuthLogin(t *testing.T) {

	t.Run("Login with DB Failure", func(t *testing.T) {

		db, _ := newSqlMockDB()

		login := authLogin(logger, db)

		req := httptest.NewRequest(http.MethodPost, "/auth/login", nil)
		req.Header.Add(echo.HeaderContentType, echo.MIMEApplicationJSON)

		rec := httptest.NewRecorder()

		c := server.NewContext(req, rec)
		c = newMockSignedContext(c)

		err := login(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})

	t.Run("Login with DB Success", func(t *testing.T) {

		db := newSqliteDB()

		assert.NoError(t, db.Create(&Profile{User: &User{Wallet: "0x123456789"}}).Error)

		login := authLogin(logger, db)

		req := httptest.NewRequest(http.MethodPost, "/auth/login", nil)
		req.Header.Add(echo.HeaderContentType, echo.MIMEApplicationJSON)

		rec := httptest.NewRecorder()
		c := server.NewContext(req, rec)
		c = newMockSignedContext(c)

		err := login(c)
		assert.NoError(t, err)
	})
}

func TestAuthLogout(t *testing.T) {

	logout := authLogout(logger, nil)

	t.Run("Logout normally with normal wallet", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)

		rec := httptest.NewRecorder()
		c := server.NewContext(req, rec)
		c = newMockSignedContext(c)

		err := logout(c)
		assert.NoError(t, err)
	})
}

func randomWallet() string {
	b := make([]byte, 32)
	var _, _ = rand.Reader.Read(b)
	return "0x" + hex.EncodeToString(b)
}

func TestCreateUserCallback(t *testing.T) {

	block := ProfileBlock{
		Username:  "Alice",
		Wallet:    randomWallet(),
		Avatar:    "null",
		Signature: "null",
		Muted:     false,
		Banned:    false,
	}

	b, _ := json.Marshal(&block)

	t.Run("Creating user with success", func(t *testing.T) {
		db := newSqliteDB()
		createUser := createUserCallback(logger, db)

		err := createUser(b)

		assert.NoError(t, err)
	})

	t.Run("Creating user with json unmarshal error", func(t *testing.T) {
		createUser := createUserCallback(logger, nil)
		err := createUser([]byte("abcd"))
		assert.Error(t, err)
	})

	t.Run("Creating user with db failure", func(t *testing.T) {
		db, _ := newSqlMockDB()
		createUser := createUserCallback(logger, db)

		err := createUser(b)

		assert.Error(t, err)

	})
}

var profile = ProfileBlock{
	Username:  "Alice",
	Wallet:    "0x123456789",
	Avatar:    "null",
	Signature: "null",
	Muted:     false,
	Banned:    false,
}

func TestUpdateUserCallback(t *testing.T) {

	type ProfileChanged struct {
		Username    string `json:"username"`
		Wallet      string `json:"wallet"`
		Avatar      string `json:"avatar"`
		Signature   string `json:"signature"`
		Credibility uint   `json:"credibility"`
		Balance     int    `json:"balance"`

		ActiveRole  uint `json:"activeRole"`
		ActiveBadge uint `json:"activeBadge"`
	}

	profileChanged := ProfileChanged{
		Username:    "Alice",
		Wallet:      "0x123456789",
		Avatar:      "avatar",
		Signature:   "signature",
		Credibility: 0,
		Balance:     0,
		ActiveRole:  0,
		ActiveBadge: 0,
	}

	t.Run("Updating user with success", func(t *testing.T) {

		db := newSqliteDB()

		updateUser := updateUserCallback(logger, db)
		createUser := createUserCallback(logger, db)

		b, _ := json.Marshal(&profile)
		err := createUser(b)

		assert.NoError(t, err)

		b, _ = json.Marshal(&profileChanged)
		err = updateUser(b)

		assert.NoError(t, err)
	})

	t.Run("Updating user with new active badge and role", func(t *testing.T) {

		db := newSqliteDB()
		assert.NoError(t, db.Create(&Badge{CID: "Qm123456789", Name: "Badge", Description: "Badge Description"}).Error)
		assert.NoError(t, db.Create(&Role{Name: "Normal User", Description: "Normal User", Privilege: 1}).Error)

		updateUser := updateUserCallback(logger, db)
		createUser := createUserCallback(logger, db)

		b, _ := json.Marshal(&profile)
		assert.NoError(t, createUser(b))

		profileChanged.ActiveBadge = 1
		profileChanged.ActiveRole = 1

		b, _ = json.Marshal(&profileChanged)

		assert.NoError(t, updateUser(b))

		profileChanged.ActiveBadge = 2
		b, _ = json.Marshal(&profileChanged)

		assert.Error(t, updateUser(b))

		profileChanged.ActiveBadge = 0
		profileChanged.ActiveRole = 3
		b, _ = json.Marshal(&profileChanged)

		assert.Error(t, updateUser(b))

	})

	t.Run("Updating user with failure", func(t *testing.T) {

		db := newSqliteDB()
		updateUser := updateUserCallback(logger, db)

		b, _ := json.Marshal(&profileChanged)
		assert.Error(t, updateUser(b))

	})

}

func TestQueryUser(t *testing.T) {

	type UserQuery struct {
		Wallet string `json:"wallet"`
	}

	t.Run("Querying User With Success", func(t *testing.T) {

		db := newSqliteDB()
		wallet := randomWallet()
		challenge := UserQuery{Wallet: wallet}

		assert.NoError(t, db.Create(&User{Wallet: wallet}).Error)

		query := queryUser(logger, db)

		req := httptest.NewRequest(http.MethodPost, "/api/user/query/view", newJsonRequest(&challenge))
		req.Header.Add(echo.HeaderContentType, echo.MIMEApplicationJSON)

		rec := httptest.NewRecorder()
		c := server.NewContext(req, rec)

		var _ = query(c)
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("Querying User With Not Found", func(t *testing.T) {

		db := newSqliteDB()

		query := queryUser(logger, db)

		challenge := UserQuery{Wallet: randomWallet()}

		req := httptest.NewRequest(http.MethodPost, "/api/user/query/view", newJsonRequest(&challenge))
		req.Header.Add(echo.HeaderContentType, echo.MIMEApplicationJSON)

		rec := httptest.NewRecorder()
		c := server.NewContext(req, rec)

		var _ = query(c)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})

	t.Run("Querying User With Unmarshal Incorrect", func(t *testing.T) {

		db := newSqliteDB()

		query := queryUser(logger, db)

		req := httptest.NewRequest(http.MethodPost, "/api/user/query/view", bytes.NewReader([]byte{1, 2, 3, 4}))
		req.Header.Add(echo.HeaderContentType, echo.MIMEApplicationJSON)

		rec := httptest.NewRecorder()
		c := server.NewContext(req, rec)

		var _ = query(c)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)

	})

}

func TestQueryProfile(t *testing.T) {

	type ProfileQuery struct {
		Wallet string `json:"wallet"`
	}

	t.Run("Querying Profile With Success", func(t *testing.T) {

		db := newSqliteDB()
		wallet := randomWallet()

		query := queryProfile(logger, db)

		challenge := ProfileQuery{Wallet: wallet}
		assert.NoError(t, db.Create(&Profile{User: &User{Wallet: wallet}}).Error)

		req := httptest.NewRequest(http.MethodPost, "/api/user/query/profile", newJsonRequest(&challenge))
		req.Header.Add(echo.HeaderContentType, echo.MIMEApplicationJSON)

		rec := httptest.NewRecorder()
		c := server.NewContext(req, rec)

		var _ = query(c)
		assert.Equal(t, http.StatusOK, rec.Code)

	})

	t.Run("Querying Profile Not Found", func(*testing.T) {

		db := newSqliteDB()
		wallet := randomWallet()

		query := queryProfile(logger, db)

		challenge := ProfileQuery{Wallet: wallet}

		req := httptest.NewRequest(http.MethodPost, "/api/user/query/profile", newJsonRequest(&challenge))
		req.Header.Add(echo.HeaderContentType, echo.MIMEApplicationJSON)

		rec := httptest.NewRecorder()
		c := server.NewContext(req, rec)

		var _ = query(c)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)

	})

	t.Run("Querying Profile Not Found", func(*testing.T) {

		db := newSqliteDB()

		query := queryProfile(logger, db)

		req := httptest.NewRequest(http.MethodPost, "/api/user/query/profile", bytes.NewReader([]byte{1, 2, 3, 4}))

		rec := httptest.NewRecorder()
		c := server.NewContext(req, rec)

		var _ = query(c)
		assert.Equal(t, http.StatusBadRequest, rec.Code)

	})
}

func TestNewProfileMiddleware(t *testing.T) {
	network := mocks.NewMockNetwork(t)
	db := newSqliteDB()
	network.EXPECT().GetContract("userprofile").Return(&client.Contract{}).Once()
	var _ = NewUserProfileMiddleware(logger, network, db)
}
