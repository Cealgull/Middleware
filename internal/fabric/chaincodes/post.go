package chaincodes

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"time"

	"github.com/Cealgull/Middleware/internal/fabric/common"
	"github.com/Cealgull/Middleware/internal/ipfs"
	. "github.com/Cealgull/Middleware/internal/models"
	"github.com/Cealgull/Middleware/internal/utils"
	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func invokeCreatePost(logger *zap.Logger, ipfs *ipfs.IPFSManager, db *gorm.DB) ChaincodeInvoke {
	return func(contract common.Contract, c echo.Context) error {

		type PostRequest struct {
			Content  string   `json:"content"`
			Images   []string `json:"images"`
			ReplyTo  string   `json:"replyTo"`
			BelongTo string   `json:"belongTo"`
		}

		postRequest := PostRequest{}

		s, _ := session.Get("session", c)
		wallet := s.Values["wallet"].(string)

		if err := c.Bind(&postRequest); err != nil {
			return c.JSON(chaincodeDeserializationError.Status(), chaincodeDeserializationError.Message())
		}

		replyPost := Post{}

		if postRequest.ReplyTo != "" {
			if err := db.Model(&Post{}).
				Where("hash = ?", postRequest.ReplyTo).First(&replyPost).Error; err != nil {
				chaincodeFieldValidationError := ChaincodeFieldValidationError{"replyTo"}
				return c.JSON(chaincodeFieldValidationError.Status(), chaincodeFieldValidationError.Message())
			}
		}

		belongTopic := Topic{}

		if err := db.Model(&Topic{}).
			Where("hash = ?", postRequest.BelongTo).First(&belongTopic).Error; err != nil {
			chaincodeFieldValidationError := ChaincodeFieldValidationError{"belongTo"}
			return c.JSON(chaincodeFieldValidationError.Status(), chaincodeFieldValidationError.Message())
		}

		ts := []byte(time.Now().String())
		ts = append(ts, []byte(wallet)...)

		hash := base64.StdEncoding.EncodeToString(sha256.New().Sum(append([]byte(postRequest.Content), ts...)))

		CID, err := ipfs.Put(bytes.NewReader([]byte(postRequest.Content)))

		if err != nil {
			return c.JSON(err.Status(), err.Message())
		}

		postBlock := PostBlock{
			Hash:     hash,
			Creator:  wallet,
			CID:      CID,
			ReplyTo:  postRequest.ReplyTo,
			BelongTo: postRequest.BelongTo,
			Assets:   postRequest.Images,
		}

		b, _ := json.Marshal(&postBlock)

		if _, err := contract.Submit("CreatePost", client.WithBytesArguments(b)); err != nil {
			chaincodeInvokeFailure := ChaincodeInvokeFailureError{"CreatePost"}
			return c.JSON(chaincodeInvokeFailure.Status(), chaincodeInvokeFailure.Message())
		}

		type PostResponse struct {
			Hash string `json:"hash"`
		}

		return c.JSON(success.Status(), &PostResponse{Hash: hash})
	}
}

func createPostCallback(logger *zap.Logger, ipfs *ipfs.IPFSManager, db *gorm.DB) ChaincodeEventCallback {

	return func(payload []byte) error {

		postBlock := PostBlock{}

		var _ = json.Unmarshal(payload, &postBlock)

		replyPost := &Post{}

		if postBlock.ReplyTo != "" {
			if err := db.Model(&Post{}).
				Where("hash = ?", postBlock.ReplyTo).First(replyPost).Error; err != nil {
				return err
			}
		} else {
			replyPost = nil
		}

		assets := utils.Map(postBlock.Assets, func(image string) *Asset {
			return &Asset{
				CreatorWallet: postBlock.Creator,
				CID:           image,
				ContentType:   "image/jpeg",
			}
		})

		data, err := ipfs.Cat(postBlock.CID)

		if err != nil {
			return err
		}

		var _ = assets

		post := Post{
			Hash:          postBlock.Hash,
			CreatorWallet: postBlock.Creator,
			Content:       string(data),

			BelongToHash: postBlock.BelongTo,
			Assets:       assets,
		}

		return db.Transaction(func(tx *gorm.DB) error {

			if err := tx.Create(&post).Error; err != nil {
				return err
			}

			if replyPost != nil {
				if err := tx.Model(&post).Association("ReplyTo").Append(replyPost); err != nil {
					return err
				}
			}

			return nil
		})
	}
}

func invokeDeletePost(logger *zap.Logger, db *gorm.DB) ChaincodeInvoke {

	return func(contract common.Contract, c echo.Context) error {

		type DeleteRequest struct {
			Hash string `json:"hash"`
		}

		deleteRequest := DeleteRequest{}

		if err := c.Bind(&deleteRequest); err != nil {
			return c.JSON(chaincodeDeserializationError.Status(), chaincodeDeserializationError.Message())
		}

		s, _ := session.Get("session", c)
		wallet := s.Values["wallet"].(string)

		post := Post{}
		if err := db.Model(&Post{}).
			Where("hash = ?", deleteRequest.Hash).First(&post).Error; err != nil {
			return err
		}

		if post.CreatorWallet != wallet {
			// TODO: return auth error
			return c.JSON(chaincodeDeserializationError.Status(), chaincodeDeserializationError.Message())
		}

		deleteBlock := DeleteBlock{
			Hash:    deleteRequest.Hash,
			Creator: wallet,
		}

		b, _ := json.Marshal(&deleteBlock)

		if _, err := contract.Submit("DeletePost", client.WithBytesArguments(b)); err != nil {
			chaincodeInvokeFailure := ChaincodeInvokeFailureError{"DeletePost"}
			return c.JSON(chaincodeInvokeFailure.Status(), chaincodeInvokeFailure.Message())
		}

		return c.JSON(success.Status(), success.Message())
	}
}

func deletePostCallback(logger *zap.Logger, db *gorm.DB) ChaincodeEventCallback {

	return func(payload []byte) error {

		deleteBlock := DeleteBlock{}

		var _ = json.Unmarshal(payload, &deleteBlock)

		return db.Transaction(func(tx *gorm.DB) error {

			post := Post{}
			if err := tx.Where("hash = ?", deleteBlock.Hash).First(&post).Error; err != nil {
				return err
			}

			return tx.Delete(&post).Error
		})
	}
}

func invokeUpdatePost(logger *zap.Logger, ipfs *ipfs.IPFSManager, db *gorm.DB) ChaincodeInvoke {
	return func(contract common.Contract, c echo.Context) error {
		type ChangePostRequest struct {
			Hash    string   `json:"hash"`
			Content string   `json:"content"`
			Images  []string `json:"images"`
		}

		postRequest := ChangePostRequest{}
		if err := c.Bind(&postRequest); err != nil {
			return c.JSON(chaincodeDeserializationError.Status(), chaincodeDeserializationError.Message())
		}

		post := Post{}
		if err := db.Model(&Post{}).
			Where("hash = ?", postRequest.Hash).First(&post).Error; err != nil {
			chaincodeFieldValidationError := ChaincodeFieldValidationError{"hash"}
			return c.JSON(chaincodeFieldValidationError.Status(), chaincodeFieldValidationError.Message())
		}

		CID, err := ipfs.Put(bytes.NewReader([]byte(postRequest.Content)))

		if err != nil {
			return c.JSON(err.Status(), err.Message())
		}

		postBlock := PostBlock{
			Hash:   postRequest.Hash,
			CID:    CID,
			Assets: postRequest.Images,
		}

		b, _ := json.Marshal(&postBlock)

		if _, err := contract.Submit("UpdatePost", client.WithBytesArguments(b)); err != nil {
			chaincodeInvokeFailure := ChaincodeInvokeFailureError{"UpdatePost"}
			return c.JSON(chaincodeInvokeFailure.Status(), chaincodeInvokeFailure.Message())
		}

		return c.JSON(success.Status(), success.Message())

	}
}

func updatePostCallback(logger *zap.Logger, ipfs *ipfs.IPFSManager, db *gorm.DB) ChaincodeEventCallback {

	return func(payload []byte) error {
		postChanged := PostBlock{}

		var _ = json.Unmarshal(payload, &postChanged)

		data, err := ipfs.Cat(postChanged.CID)

		if err != nil {
			return err
		}

		assets := utils.FilterMap(postChanged.Assets, func(image string) *Asset {
			return &Asset{
				CreatorWallet: postChanged.Creator,
				CID:           image,
				ContentType:   "image/jpeg",
			}
		}, func(image string) bool {
			return image != NONE
		})

		post := Post{}

		var _ = assets

		return db.Transaction(func(tx *gorm.DB) error {

			if err := tx.Model(&Post{}).
				Where("hash = ?", postChanged.Hash).First(&post).Error; err != nil {
				return err
			}

			if len(postChanged.Assets) != 0 {
				if err := tx.Model(&post).Association("Assets").Replace(&assets); err != nil {
					return err
				}
			}

			return tx.Model(&post).
				Updates(&Post{Content: string(data)}).Error

		})
	}
}

func invokeUpvotePost(logger *zap.Logger, db *gorm.DB) ChaincodeInvoke {
	return func(contract common.Contract, c echo.Context) error {
		type UpvoteRequest struct {
			Hash string `json:"hash"`
		}

		upvoteRequest := UpvoteRequest{}
		if err := c.Bind(&upvoteRequest); err != nil {
			return c.JSON(chaincodeDeserializationError.Status(), chaincodeDeserializationError.Message())
		}

		post := Post{}
		if err := db.Model(&Post{}).
			Where("hash = ?", upvoteRequest.Hash).First(&post).Error; err != nil {
			return err
		}

		s, _ := session.Get("session", c)
		wallet := s.Values["wallet"].(string)

		upvoteBlock := UpvoteBlock{
			Hash:    upvoteRequest.Hash,
			Creator: wallet,
		}
		b, _ := json.Marshal(&upvoteBlock)

		if _, err := contract.Submit("UpvotePost", client.WithBytesArguments(b)); err != nil {
			chaincodeInvokeFailure := ChaincodeInvokeFailureError{"UpvotePost"}
			return c.JSON(chaincodeInvokeFailure.Status(), chaincodeInvokeFailure.Message())
		}

		return c.JSON(success.Status(), success.Message())

	}
}

func upvotePostCallback(logger *zap.Logger, db *gorm.DB) ChaincodeEventCallback {

	return func(payload []byte) error {

		upvoteBlock := UpvoteBlock{}

		var _ = json.Unmarshal(payload, &upvoteBlock)

		post := Post{}

		if err := db.Model(&Post{}).
			Preload("Upvotes").
      Preload("Downvotes").
			Where("hash = ?", upvoteBlock.Hash).First(&post).Error; err != nil {
			return err
		}

		return db.Transaction(func(tx *gorm.DB) error {

			upvote := Upvote{
				CreatorWallet: upvoteBlock.Creator,
			}

			for _, u := range post.Upvotes {
				if u.CreatorWallet == upvote.CreatorWallet {
					tx.Model(&post).Association("Upvotes").Delete(u)
					return nil
				}
			}

			for _, d := range post.Downvotes {
				if d.CreatorWallet == upvote.CreatorWallet {
					tx.Model(&post).Association("Downvotes").Delete(d)
					break
				}
			}

			var _ = tx.Model(&post).Association("Upvotes").Append(&upvote)

			return nil
		})
	}
}

func invokeDownvotePost(logger *zap.Logger, db *gorm.DB) ChaincodeInvoke {
	return func(contract common.Contract, c echo.Context) error {
		type DownvoteRequest struct {
			Hash string `json:"hash"`
		}

		downvoteRequest := DownvoteRequest{}
		if err := c.Bind(&downvoteRequest); err != nil {
			return c.JSON(chaincodeDeserializationError.Status(), chaincodeDeserializationError.Message())
		}

		post := Post{}
		if err := db.Model(&Post{}).
			Where("hash = ?", downvoteRequest.Hash).First(&post).Error; err != nil {
			return err
		}

		s, _ := session.Get("session", c)
		wallet := s.Values["wallet"].(string)

		downvoteBlock := DownvoteBlock{
			Hash:    downvoteRequest.Hash,
			Creator: wallet,
		}
		b, _ := json.Marshal(&downvoteBlock)

		if _, err := contract.Submit("DownvotePost", client.WithBytesArguments(b)); err != nil {
			chaincodeInvokeFailure := ChaincodeInvokeFailureError{"DownvotePost"}
			return c.JSON(chaincodeInvokeFailure.Status(), chaincodeInvokeFailure.Message())
		}

		return c.JSON(success.Status(), success.Message())

	}
}

func downvotePostCallback(logger *zap.Logger, db *gorm.DB) ChaincodeEventCallback {

	return func(payload []byte) error {

		downvoteBlock := DownvoteBlock{}

		var _ = json.Unmarshal(payload, &downvoteBlock)

		post := Post{}

		if err := db.Model(&Post{}).
      Preload("Upvotes").
			Preload("Downvotes").
			Where("hash = ?", downvoteBlock.Hash).First(&post).Error; err != nil {
			return err
		}

		return db.Transaction(func(tx *gorm.DB) error {

			downvote := Downvote{
				CreatorWallet: downvoteBlock.Creator,
			}

			for _, d := range post.Downvotes {
				if d.CreatorWallet == downvote.CreatorWallet {
					tx.Model(&post).Association("Downvotes").Delete(d)
					return nil
				}
			}

			for _, u := range post.Upvotes {
				if u.CreatorWallet == downvote.CreatorWallet {
					tx.Model(&post).Association("Upvotes").Delete(u)
					break
				}
			}

			var _ = tx.Model(&post).Association("Downvotes").Append(&downvote)
			return nil
		})
	}
}

func queryPostsList(logger *zap.Logger, db *gorm.DB) ChaincodeQuery {
	return func(c echo.Context) error {

		type QueryRequest struct {
			PageOrdinal int    `json:"pageOrdinal"`
			PageSize    int    `json:"pageSize"`
			BelongTo    string `json:"belongTo"`
			Creator     string `json:"creator"`
		}

		q := QueryRequest{}

		if c.Bind(&q) != nil {
			return c.JSON(chaincodeDeserializationError.Status(), chaincodeDeserializationError.Message())
		}

		if q.PageOrdinal <= 0 || q.PageSize <= 0 {
			return c.JSON(chaincodeQueryParameterError.Status(), chaincodeQueryParameterError.Message())
		}

		posts := []Post{}

		err := db.Transaction(func(tx *gorm.DB) error {

			tx = tx.Model(&Post{}).
        Preload("Creator").
				Preload("ReplyTo").
				Preload("ReplyTo.Creator").
				Preload("ReplyTo.Assets").
        Preload("Upvotes").
        Preload("Downvotes").
        Preload("BelongTo").
				Scopes(paginate(q.PageOrdinal, q.PageSize))

			if q.Creator != "" {
				tx = tx.Where("creator_wallet = ?", q.Creator)
			}

			if q.BelongTo != "" {
				tx = tx.
					Where("belong_to_hash = ?", q.BelongTo)
			}

			tx = tx.Where("deleted_at IS NULL")

			if err := tx.Find(&posts).Error; err != nil {
				return err
			}

			return nil
		})

		if err != nil {
			return c.JSON(chaincodeInternalError.Status(), chaincodeInternalError.Message())
		}

		return c.JSON(success.Status(), posts)

	}
}

func NewPostChaincodeMiddleware(logger *zap.Logger, net common.Network, ipfs *ipfs.IPFSManager, db *gorm.DB) *ChaincodeMiddleware {
	return NewChaincodeMiddleware(logger, net, net.GetContract("post"),

		WithChaincodeHandler("create", "CreatePost", invokeCreatePost(logger, ipfs, db), createPostCallback(logger, ipfs, db)),
		WithChaincodeHandler("update", "UpdatePost", invokeUpdatePost(logger, ipfs, db), updatePostCallback(logger, ipfs, db)),

		WithChaincodeHandler("upvote", "UpvotePost", invokeUpvotePost(logger, db), upvotePostCallback(logger, db)),
		WithChaincodeHandler("downvote", "DownvotePost", invokeDownvotePost(logger, db), downvotePostCallback(logger, db)),
		WithChaincodeHandler("delete", "DeletePost", invokeDeletePost(logger, db), deletePostCallback(logger, db)),

		WithChaincodeQuery("list", queryPostsList(logger, db)),
	)
}
