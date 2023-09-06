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

		images := make([]string, len(postRequest.Images))

		for i, imageb64 := range postRequest.Images {

			data, err := base64.StdEncoding.DecodeString(imageb64)

			if err != nil {
				return c.JSON(chaincodeBase64DecodeError.Status(), chaincodeBase64DecodeError.Message())
			}

			if cid, err := ipfs.Put(bytes.NewReader(data)); err != nil {
				return c.JSON(err.Status(), err.Message())
			} else {
				images[i] = cid
			}

		}

		postBlock := PostBlock{
			Hash:     hash,
			Creator:  wallet,
			CID:      CID,
			ReplyTo:  postRequest.ReplyTo,
			BelongTo: postRequest.BelongTo,
			Assets:   images,
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

		images := make([]string, len(postRequest.Images))

		for i, imageb64 := range postRequest.Images {

			data, err := base64.StdEncoding.DecodeString(imageb64)

			if err != nil {
				return c.JSON(chaincodeBase64DecodeError.Status(), chaincodeBase64DecodeError.Message())
			}

			if cid, err := ipfs.Put(bytes.NewReader(data)); err != nil {
				return c.JSON(err.Status(), err.Message())
			} else {
				images[i] = cid
			}
		}

		postBlock := PostBlock{
			Hash:   postRequest.Hash,
			CID:    CID,
			Assets: images,
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

		assets := utils.Map(postChanged.Assets, func(image string) *Asset {
			return &Asset{
				CreatorWallet: postChanged.Creator,
				CID:           image,
				ContentType:   "image/jpeg",
			}
		})

		post := Post{}

		var _ = assets

		return db.Transaction(func(tx *gorm.DB) error {

			if err := tx.Model(&Post{}).
				Where("hash = ?", postChanged.Hash).First(&post).Error; err != nil {
				return err
			}

      return tx.Model(&post).
        Updates(&Post{Content: string(data), Assets: assets}).Error

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
				Preload("ReplyTo").
				Preload("ReplyTo.Creator").
				Preload("ReplyTo.Assets").
				Scopes(paginate(q.PageOrdinal, q.PageSize))

			if q.Creator != "" {
				tx = tx.Where("creator_wallet = ?", q.Creator)
			}

			if q.BelongTo != "" {
				tx = tx.
					Where("belong_to_hash = ?", q.BelongTo)
			}

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

		WithChaincodeQuery("list", queryPostsList(logger, db)),
	)
}
