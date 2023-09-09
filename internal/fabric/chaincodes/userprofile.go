package chaincodes

import (
	"encoding/json"
	"errors"
	"net/http"

	"time"

	"github.com/Cealgull/Middleware/internal/fabric/common"
	. "github.com/Cealgull/Middleware/internal/models"
	"github.com/Cealgull/Middleware/internal/utils"
	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	// "gorm.io/hints"
)

func invokeCreateUser(logger *zap.Logger, db *gorm.DB) ChaincodeInvoke {
	return func(contract common.Contract, c echo.Context) error {

		s, _ := session.Get("session", c)

		wallet := s.Values["wallet"].(string)

		if err := db.Model(&User{}).Where(&User{Wallet: wallet}).First(&User{}).Error; err == nil {
			chaincodeDuplicateError := ChaincodeDuplicatedError{"User"}
			return c.JSON(chaincodeDuplicateError.Status(), chaincodeDuplicateError.Message())
		}

		block := ProfileBlock{
			Username:  "Alice",
			Wallet:    wallet,
			Signature: "Alice's signature",
			Muted:     false,
			Banned:    false,
		}

		logger.Debug("Invoke Creating User", zap.String("wallet", block.Wallet))

		b, _ := json.Marshal(&block)

		if _, err := contract.Submit("CreateUser", client.WithBytesArguments(b)); err != nil {
			chaincodeInvokeFailure := ChaincodeInvokeFailureError{"CreateUser"}
			return c.JSON(chaincodeInvokeFailure.Status(), chaincodeInvokeFailure.Message())
		}

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

		return c.JSON(success.Status(), &profile)

	}
}

func invokeUpdateUser(logger *zap.Logger, db *gorm.DB) ChaincodeInvoke {

	return func(contract common.Contract, c echo.Context) error {

		type ProfileChanged struct {
			Username  string `json:"username"`
			Wallet    string `json:"wallet"`
			Avatar    string `json:"avatar"`
			Signature string `json:"signature"`

			ActiveRole  string `json:"activeRole"`
			ActiveBadge string `json:"activeBadge"`
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

			if profile.ActiveRole != "" {
				if !utils.Contains(utils.Map(userProfile.RoleRelationsAssigned, func(r *RoleRelation) string {
					return r.RoleName
				}), profile.ActiveRole) {
					return errors.New("The user does not the role")
				}
			}

			if profile.ActiveBadge != "" {
				if !utils.Contains(utils.Map(userProfile.BadgeRelationsReceived, func(r *BadgeRelation) string {
					return r.BadgeName
				}), profile.ActiveBadge) {
					return errors.New("user does not have the role or badge")
				}
			}

			return nil
		})

		if err != nil {
			return c.JSON(chaincodeInternalError.Status(), chaincodeInternalError.Message())
		}

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

			ActiveRole  string `json:"activeRole"`
			ActiveBadge string `json:"activeBadge"`
		}

		profileChanged := ProfileChanged{}

		var _ = json.Unmarshal(payload, &profileChanged)

		return db.Transaction(func(tx *gorm.DB) error {

			user := User{
				Username: profileChanged.Username,
				Avatar:   profileChanged.Avatar,
			}

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

			if len(profileChanged.ActiveBadge) != 0 {
				if err := tx.Model(prevProfile.User).Association("ActiveBadgeRelation").
					Append(&BadgeRelation{BadgeName: profileChanged.ActiveBadge}); err != nil {
					return err
				}
			}

			if len(profileChanged.ActiveRole) != 0 {
				if err := tx.Model(prevProfile.User).Association("ActiveRoleRelation").
					Append(&RoleRelation{RoleName: profileChanged.ActiveRole}); err != nil {
					return err
				}
			}

			return nil
		})
	}
}

func authLogin(logger *zap.Logger, db *gorm.DB) ChaincodeCustom {
	return func(contract common.Contract, c echo.Context) error {
		s, _ := session.Get("session", c)
		wallet := s.Values["wallet"].(string)

		profile := Profile{}

		if err := db.
			Preload(clause.Associations).
			Preload("User.ActiveBadgeRelation").
			Preload("User.ActiveRoleRelation").
			Where("user_wallet = ?", wallet).
			First(&profile).Error; err == nil {
			return c.JSON(http.StatusOK, &profile)
		} else if errors.Is(err, gorm.ErrRecordNotFound) {
			return invokeCreateUser(logger, db)(contract, c)
		} else {
			return c.JSON(chaincodeInternalError.Status(), chaincodeInternalError.Message())
		}

	}
}

func authLogout(logger *zap.Logger, db *gorm.DB) ChaincodeCustom {
	return func(contract common.Contract, c echo.Context) error {
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

func queryStatistics(logger *zap.Logger, db *gorm.DB) ChaincodeQuery {
	return func(c echo.Context) error {

		type StatQuery struct {
			Wallet string `json:"wallet"`
		}

		q := StatQuery{}

		if c.Bind(&q) != nil {
			chaincodeDeserializationError := ChaincodeDeserializationError{}
			return c.JSON(chaincodeDeserializationError.Status(), chaincodeDeserializationError.Message())
		}

		type StatResponse struct {
			UpvotesGranted  int       `json:"upvotesGranted"`
			UpvotesReceived int       `json:"upvotesRecieved"`
			TopicsCreated   int       `json:"topicsCreated"`
			PostsCreated    int       `json:"postsCreated"`
			RegisterDate    time.Time `json:"registerDate"`
		}

		if db.Model(&User{}).Where("wallet = ?", q.Wallet).First(&User{}).Error != nil {
			chaincodeNotFoundError := ChaincodeNotFoundError{}
			return c.JSON(chaincodeNotFoundError.Status(), chaincodeNotFoundError.Message())
		}

		r := &StatResponse{}

		var _ = db.Transaction(func(tx *gorm.DB) error {

			upvotesGrantedQuery := tx.Table("upvotes").Select("COUNT(*) as upvotes_granted").Where("creator_wallet = ?", q.Wallet)
			postsCreatedQuery := tx.Table("posts").Select("COUNT(*) as posts_created").Where("creator_wallet = ?", q.Wallet)
			topicsCreatedQuery := tx.Table("topics").Select("COUNT(*) as topics_created").Where("creator_wallet = ?", q.Wallet)
			registerDateQuery := tx.Table("users").Select("created_at as register_date").Where("wallet = ?", q.Wallet)

			upvotesPostsReceivedQuery := tx.Table("upvotes").
				Select("COUNT (*) as u1").
				Where("owner_type = ?", "posts").
				Joins("inner join (?) as p on p.id = upvotes.owner_id",
					db.Model(&Post{}).Where("creator_wallet = ?", q.Wallet))

			upvotesTopicReceivedQuery := tx.Table("upvotes").
				Select("COUNT (*) as u2").
				Where("owner_type = ?", "topics").
				Joins("inner join (?) as t on t.id = upvotes.owner_id",
					db.Model(&Topic{}).Where("creator_wallet = ?", q.Wallet))

			upvotesReceivedQuery := tx.Table("(?) as u1, (?) as u2", upvotesPostsReceivedQuery, upvotesTopicReceivedQuery).Select("(u1 + u2) as upvotes_received")

			return tx.Table("(?) as upvoted_granted, (?) as upvotes_received, (?) as topics_created, (?) as posts_created, (?) as register_date",
				upvotesGrantedQuery,
				upvotesReceivedQuery,
				topicsCreatedQuery,
				postsCreatedQuery,
				registerDateQuery).Scan(r).Error

		})

		return c.JSON(success.Status(), r)
	}

}

func NewUserProfileMiddleware(logger *zap.Logger, net common.Network, db *gorm.DB) *ChaincodeMiddleware {

	return NewChaincodeMiddleware(logger, net, net.GetContract("userprofile"),

		WithChaincodeHandler("create", "CreateUser", invokeCreateUser(logger, db), createUserCallback(logger, db)),
		WithChaincodeHandler("update", "UpdateUser", invokeUpdateUser(logger, db), updateUserCallback(logger, db)),

		WithChaincodeQueryPost("profile", queryProfile(logger, db)),
		WithChaincodeQueryPost("view", queryUser(logger, db)),
		WithChaincodeQueryPost("statistics", queryStatistics(logger, db)),

		WithChaincodeCustom("/auth/login", authLogin(logger, db)),
		WithChaincodeCustom("/auth/logout", authLogin(logger, db)),
	)
}
