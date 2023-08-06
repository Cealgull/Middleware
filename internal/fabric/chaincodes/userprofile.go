package chaincodes

import (
	"encoding/json"
	"net/http"

	"github.com/Cealgull/Middleware/internal/models"
	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func invokeCreateUser(logger *zap.Logger) ChaincodeInvoke {
	return func(contract *client.Contract, c echo.Context) error {

		s, _ := session.Get("session", c)

		wallet := s.Values["wallet"].(string)

		profile := models.ProfileBlock{
			Username:  "Alice",
			Wallet:    wallet,
			Signature: "Alice's signature",
			Muted:     false,
			Banned:    false,
		}

		logger.Debug("Invoke Creating User", zap.String("wallet", profile.Wallet))

		b, _ := json.Marshal(&profile)

		b, err := contract.Submit("CreateUser",
			client.WithBytesArguments(b))

		if err != nil {
			chaincodeInvokeFailure := ChaincodeInvokeFailureError{"CreateUser"}
			return c.JSON(chaincodeInvokeFailure.Status(), chaincodeInvokeFailure.Message())
		}

		return c.JSON(success.Status(), success.Message())

	}
}

func invokeUpdateUser(logger *zap.Logger) ChaincodeInvoke {

	return func(contract *client.Contract, c echo.Context) error {

		type ProfileChanged struct {
			Username  string `json:"username"`
			Wallet    string `json:"wallet"`
			Avatar    string `json:"avatar"`
			Signature string `json:"signature"`

			ActiveRole  uint `json:"activeRole"`
			ActiveBadge uint `json:"activeBadge"`
		}

		profile := ProfileChanged{}

		if err := c.Bind(&profile); err != nil {
			return c.JSON(chaincodeDeserializationError.Status(), chaincodeDeserializationError.Message())
		}

		s, _ := session.Get("session", c)

		profile.Wallet = s.Values["wallet"].(string)

		b, err := json.Marshal(&profile)

		if err != nil {
			return err
		}

		_, err = contract.Submit("UpdateUser", client.WithBytesArguments(b))

		if err != nil {
			chaincodeInvokeFailure := ChaincodeInvokeFailureError{"UpdateUser"}
			return c.JSON(chaincodeInvokeFailure.Status(), chaincodeInvokeFailure.Message())
		}

		return c.JSON(success.Status(), success.Message())

	}
}

func createUserCallback(logger *zap.Logger, db *gorm.DB) ChaincodeEventCallback {
	return func(payload []byte) error {

		block := models.ProfileBlock{}

		if err := json.Unmarshal(payload, &block); err != nil {
			return err
		}

		logger.Info("Receiving create user",
			zap.String("username", block.Username),
			zap.String("wallet", block.Wallet))

		return db.Transaction(func(tx *gorm.DB) error {

			user := models.User{
				Username:    block.Username,
				Wallet:      block.Wallet,
				Avatar:      block.Avatar,
				Muted:       block.Muted,
				Banned:      block.Banned,
				ActiveRole:  nil,
				ActiveBadge: nil,
			}

			profile := models.Profile{
				Signature:   block.Signature,
				Balance:     block.Balance,
				Credibility: block.Credibility,
				User:        &user,
			}

			if err := tx.Create(&profile).Error; err != nil {
				return err
			}
			return nil
		})
	}
}

func updateUserCallback(logger *zap.Logger, db *gorm.DB) ChaincodeEventCallback {
	return func(payload []byte) error {

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

		profileChanged := ProfileChanged{}

		var _ = json.Unmarshal(payload, &profileChanged)

		return db.Transaction(func(tx *gorm.DB) error {

			user := models.User{
				Username: profileChanged.Username,
				Avatar:   profileChanged.Avatar,
			}

			profile := models.Profile{
				Signature:   profileChanged.Signature,
				Credibility: profileChanged.Credibility,
				Balance:     profileChanged.Balance,
				User:        &user,
			}

			if err := tx.Model(&models.Profile{}).Where("user_wallet = ?", profileChanged.Wallet).Updates(&profile).Error; err != nil {
				return err
			}

			if profileChanged.ActiveBadge != 0 {
				db.
					Model(&models.User{}).
					Where("wallet = ?", profileChanged.Wallet).
					Association("ActiveBadge").
					Replace(&models.Badge{ID: profileChanged.ActiveBadge})
			}

			if profileChanged.ActiveRole != 0 {
				db.
					Model(&models.User{}).
					Where("wallet = ?", profileChanged.Wallet).
					Association("ActiveRole").
					Replace(&models.Badge{ID: profileChanged.ActiveRole})
			}

			return nil

		})
	}
}

func authLogin(logger *zap.Logger, db *gorm.DB) ChaincodeCustom {
	return func(c echo.Context) error {
		s, _ := session.Get("session", c)
		wallet := s.Values["wallet"].(string)

		profile := models.Profile{}

		if err := db.
			Preload(clause.Associations).
			Preload("User.ActiveBadge").
			Preload("User.ActiveRole").
			Where("user_wallet = ?", wallet).
			First(&profile).Error; err != nil {

			return c.JSON(chaincodeInternalError.Status(), chaincodeInternalError.Message())
		}

		return c.JSON(http.StatusOK, &profile)
	}
}

func authLogout(logger *zap.Logger, db *gorm.DB) ChaincodeCustom {
	return func(c echo.Context) error {
		s, _ := session.Get("session", c)

		s.Options.MaxAge = -1

		if err := s.Save(c.Request(), c.Response()); err != nil {
			return c.JSON(chaincodeDeserializationError.Status(), chaincodeDeserializationError.Message())
		}

		return c.JSON(success.Status(), success.Message())

	}
}

func queryProfile(logger *zap.Logger, db *gorm.DB) ChaincodeQuery {
	return func(c echo.Context) error {

		type ProfileQuery struct {
			Wallet string `json:"wallet"`
		}

		profileQuery := ProfileQuery{}

		if err := c.Bind(&profileQuery); err != nil {
			return c.JSON(chaincodeDeserializationError.Status(), chaincodeDeserializationError.Message())
		}

		profile := models.Profile{}

		if err := db.
			Preload(clause.Associations).
			Preload("User.ActiveBadge").
			Preload("User.ActiveRole").
			Where("user_wallet = ?", profileQuery.Wallet).
			First(&models.Profile{}).Error; err != nil {
			return c.JSON(chaincodeInternalError.Status(), chaincodeInternalError.Message())
		}

		return c.JSON(success.Status(), &profile)
	}
}

func queryUser(logger *zap.Logger, db *gorm.DB) ChaincodeQuery {
	return func(c echo.Context) error {

		type UserQuery struct {
			Wallet string `json:"wallet"`
		}

		userQuery := UserQuery{}

		if err := c.Bind(&userQuery); err != nil {
			return c.JSON(chaincodeInternalError.Status(), chaincodeInternalError.Message())
		}

		user := models.User{}

		if err := db.
			Preload(clause.Associations).
			Preload("ActiveBadge").
			Preload("ActiveRole").
			Where("wallet = ?", userQuery.Wallet).
			First(&user).Error; err != nil {
			return c.JSON(chaincodeInternalError.Status(), chaincodeInternalError.Message())
		}

		return c.JSON(success.Status(), &user)
	}
}

func NewUserProfileMiddleware(logger *zap.Logger, net *client.Network, db *gorm.DB) *ChaincodeMiddleware {

	return NewChaincodeMiddleware(logger, net, "userprofile",

		WithChaincodeHandler("create", "CreateUser", invokeCreateUser(logger), createUserCallback(logger, db)),
		WithChaincodeHandler("update", "UpdateUser", invokeUpdateUser(logger), updateUserCallback(logger, db)),
		WithChaincodeQuery("profile", queryProfile(logger, db)),
		WithChaincodeQuery("user", queryUser(logger, db)),

		WithChaincodeCustom("/auth/login", authLogin(logger, db)),
		WithChaincodeCustom("/auth/logout", authLogin(logger, db)),
	)
}
