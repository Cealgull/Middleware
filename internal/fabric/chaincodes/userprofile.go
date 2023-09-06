package chaincodes

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/Cealgull/Middleware/internal/fabric/common"
	. "github.com/Cealgull/Middleware/internal/models"
	"github.com/Cealgull/Middleware/internal/utils"
	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func invokeCreateUser(logger *zap.Logger) ChaincodeInvoke {
	return func(contract common.Contract, c echo.Context) error {

		s, _ := session.Get("session", c)

		wallet := s.Values["wallet"].(string)

		profile := ProfileBlock{
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

func invokeUpdateUser(logger *zap.Logger, db *gorm.DB) ChaincodeInvoke {

	return func(contract common.Contract, c echo.Context) error {

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

		err := db.Transaction(func(tx *gorm.DB) error {
			var userProfile Profile
			if err := tx.Model(&Profile{}).
				Preload("RoleRelationsAssigned").
				Preload("BadgeRelationsReceived").
				Where("user_wallet = ?", profile.Wallet).
				First(&userProfile).Error; err != nil {
				return err
			}
			if !utils.Contains(utils.Map(userProfile.RoleRelationsAssigned, func(r *RoleRelation) uint {
				return r.RoleID
			}), profile.ActiveRole) || !utils.Contains(utils.Map(userProfile.BadgeRelationsReceived, func(r *BadgeRelation) uint {
				return r.BadgeID
			}), profile.ActiveBadge) {
				return errors.New("User does not have the role or badge")
			}
			return nil
		})

		b, _ := json.Marshal(&profile)

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

		block := ProfileBlock{}

		if err := json.Unmarshal(payload, &block); err != nil {
			return err
		}

		logger.Info("Receiving create user",
			zap.String("username", block.Username),
			zap.String("wallet", block.Wallet))

		return db.Transaction(func(tx *gorm.DB) error {

			user := User{
				Username:            block.Username,
				Wallet:              block.Wallet,
				Avatar:              block.Avatar,
				Muted:               block.Muted,
				Banned:              block.Banned,
				ActiveRoleRelation:  nil,
				ActiveBadgeRelation: nil,
			}

			profile := Profile{
				Signature:   block.Signature,
				Balance:     block.Balance,
				Credibility: block.Credibility,
				User:        &user,
			}

			var _ = tx.Create(&profile).Error

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

			user := User{
				Username: profileChanged.Username,
				Avatar:   profileChanged.Avatar,
			}
      
      fmt.Println(user)

			profile := Profile{
				Signature:   profileChanged.Signature,
				Credibility: profileChanged.Credibility,
				Balance:     profileChanged.Balance,
			}

			prevProfile := Profile{}

			if err := tx.Preload("User").Model(&Profile{}).
				Where("user_wallet = ?", profileChanged.Wallet).First(&prevProfile).Error; err != nil {
				return err
			}
      
      profile.ID = prevProfile.ID
      user.ID = prevProfile.User.ID

			var _ = tx.Updates(&profile).Error
      var _ = tx.Updates(&user).Error

			if profileChanged.ActiveBadge != 0 {
				if err := tx.Model(prevProfile.User).Association("ActiveBadgeRelation").
					Append(&BadgeRelation{BadgeID: profileChanged.ActiveBadge}); err != nil {
					return err
				}
			}

			if profileChanged.ActiveRole != 0 {
				if err := tx.Model(prevProfile.User).Association("ActiveRoleRelation").
					Append(&RoleRelation{RoleID: profileChanged.ActiveRole}); err != nil {
					return err
				}
			}

			return nil
		})
	}
}

func authLogin(logger *zap.Logger, db *gorm.DB) ChaincodeCustom {
	return func(c echo.Context) error {
		s, _ := session.Get("session", c)
		wallet := s.Values["wallet"].(string)

		profile := Profile{}

		if err := db.
			Preload(clause.Associations).
			Preload("User.ActiveBadgeRelation").
			Preload("User.ActiveRoleRelation").
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

		var _ = s.Save(c.Request(), c.Response())

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

		profile := Profile{}

		if err := db.
			Preload(clause.Associations).
			Preload("User.ActiveBadgeRelation").
			Preload("User.ActiveRoleRelation").
			Where("user_wallet = ?", profileQuery.Wallet).
			First(&profile).Error; err != nil {
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

		user := User{}

		if err := db.
			Preload(clause.Associations).
			Preload("ActiveBadgeRelation").
			Preload("ActiveRoleRelation").
			Where("wallet = ?", userQuery.Wallet).
			First(&user).Error; err != nil {
			return c.JSON(chaincodeInternalError.Status(), chaincodeInternalError.Message())
		}

		return c.JSON(success.Status(), &user)
	}
}

func NewUserProfileMiddleware(logger *zap.Logger, net common.Network, db *gorm.DB) *ChaincodeMiddleware {

	return NewChaincodeMiddleware(logger, net, net.GetContract("userprofile"),

		WithChaincodeHandler("create", "CreateUser", invokeCreateUser(logger), createUserCallback(logger, db)),
		WithChaincodeHandler("update", "UpdateUser", invokeUpdateUser(logger, db), updateUserCallback(logger, db)),

		WithChaincodeQuery("profile", queryProfile(logger, db)),
		WithChaincodeQuery("view", queryUser(logger, db)),

		WithChaincodeCustom("/auth/login", authLogin(logger, db)),
		WithChaincodeCustom("/auth/logout", authLogin(logger, db)),
	)
}
