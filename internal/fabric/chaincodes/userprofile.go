package chaincodes

import (
	"encoding/json"

	"github.com/Cealgull/Middleware/internal/models"
	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func InvokeCreateUser(logger *zap.Logger) ChaincodeInvoke {
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

		b, err := json.Marshal(&profile)

		if err != nil {
			return err
		}

		b, err = contract.Submit("CreateUser",
			client.WithBytesArguments(b))

		return err
	}
}

func InvokeUpdateUser(logger *zap.Logger) ChaincodeInvoke {

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
			return err
		}

		s, _ := session.Get("session", c)

		profile.Wallet = s.Values["wallet"].(string)

		b, err := json.Marshal(&profile)

		if err != nil {
			return err
		}

		_, err = contract.Submit("UpdateUser", client.WithBytesArguments(b))

		return err

	}
}

func CreateUserCallback(logger *zap.Logger, db *gorm.DB) ChaincodeEventCallback {
	return func(payload []byte) error {

		block := models.ProfileBlock{}
		var _ = json.Unmarshal(payload, &block)

		logger.Info("Receiving create user",
			zap.String("username", block.Username),
			zap.String("wallet", block.Wallet))

		return db.Transaction(func(tx *gorm.DB) error {

			user := models.User{
				Wallet:      block.Wallet,
				Username:    block.Username,
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

func UpdateUserCallback(logger *zap.Logger, db *gorm.DB) ChaincodeEventCallback {
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

			if profileChanged.ActiveBadge != 0 {
				user.ActiveBadge = &models.Badge{ID: profileChanged.ActiveBadge}
			}

			if profileChanged.ActiveRole != 0 {
				user.ActiveRole = &models.Role{ID: profileChanged.ActiveRole}
			}

			if err := tx.Model(&models.Profile{}).Where("wallet = ?", profileChanged.Wallet).Updates(&profile).Error; err != nil {
				return err
			}

			return nil

		})
	}
}

func NewUserProfileMiddleware(logger *zap.Logger, db *gorm.DB, net *client.Network) *ChaincodeMiddleware {

	return NewChaincodeMiddleware(logger, net, "userprofile",
		WithChaincodeHandler("create", "CreateUser",
			InvokeCreateUser(logger),
			CreateUserCallback(logger, db)),
		WithChaincodeHandler("update", "UpdateUser",
			InvokeUpdateUser(logger),
			UpdateUserCallback(logger, db)))
}
