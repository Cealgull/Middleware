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
		if err := db.Model(&Post{}).
			Where("hash = ?", postRequest.ReplyTo).First(&replyPost).Error; err != nil {
			return err
		}
		belongTopic := Topic{}
		if err := db.Model(&Topic{}).
			Where("hash = ?", postRequest.BelongTo).First(&belongTopic).Error; err != nil {
			return err
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
			CreateAt: time.Now(),
			UpdateAt: time.Now(),
			ReplyTo:  postRequest.ReplyTo,
			BelongTo: postRequest.BelongTo,
			Assets:   images,
		}

		b, _ := json.Marshal(&postBlock)

		if _, err := contract.Submit("CreatePost", client.WithBytesArguments(b)); err != nil {
			chaincodeInvokeFailure := ChaincodeInvokeFailureError{"CreatePost"}
			return c.JSON(chaincodeInvokeFailure.Status(), chaincodeInvokeFailure.Message())
		}

		return c.JSON(success.Status(), success.Message())
	}
}

func createPostCallback(logger *zap.Logger, ipfs *ipfs.IPFSManager, db *gorm.DB) ChaincodeEventCallback {

	return func(payload []byte) error {

		postBlock := PostBlock{}

		var _ = json.Unmarshal(payload, &postBlock)

		replyPost := Post{}
		if err := db.Model(&Post{}).
			Where("hash = ?", postBlock.ReplyTo).First(&replyPost).Error; err != nil {
			return err
		}
		belongTopic := Topic{}
		if err := db.Model(&Topic{}).
			Where("hash = ?", postBlock.BelongTo).First(&belongTopic).Error; err != nil {
			return err
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

		return db.Transaction(func(tx *gorm.DB) error {

			post := Post{
				Hash:          postBlock.Hash,
				CreatorWallet: postBlock.Creator,
				Content:       string(data),
				CreateAt:      postBlock.CreateAt,
				UpdateAt:      postBlock.UpdateAt,

				BelongTo: &belongTopic,
				ReplyTo:  &replyPost,
				Assets:   assets,
			}

			if err := tx.Create(&post).Error; err != nil {
				return err
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
			Assets  []string `json:"assets"`
			Type    string   `json:"type"`
		}

		postRequest := ChangePostRequest{}
		if err := c.Bind(&postRequest); err != nil {
			return c.JSON(chaincodeDeserializationError.Status(), chaincodeDeserializationError.Message())
		}

		post := Post{}
		if err := db.Model(&Post{}).
			Where("hash = ?", postRequest.Hash).First(&post).Error; err != nil {
			return err
		}

		CID, err := ipfs.Put(bytes.NewReader([]byte(postRequest.Content)))

		if err != nil {
			return c.JSON(err.Status(), err.Message())
		}

		images := make([]string, len(postRequest.Assets))

		for i, imageb64 := range postRequest.Assets {

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
			Hash:     postRequest.Hash,
			CID:      CID,
			UpdateAt: time.Now(),
			Assets:   images,
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

		post := Post{}
		if err := db.Model(&Post{}).
			Where("hash = ?", postChanged.Hash).First(&post).Error; err != nil {
			return err
		}

		data, err := ipfs.Cat(postChanged.CID)
		if err != nil {
			return err
		}

		assets := utils.Map(postChanged.Assets, func(image string) *Asset {
			return &Asset{
				CreatorWallet: post.CreatorWallet,
				CID:           image,
				ContentType:   "image/jpeg",
			}
		})

		var _ = assets

		return db.Transaction(func(tx *gorm.DB) error {
			tx.Model(&post).Select("content", "update_at", "assets").Updates(map[string]interface{}{"content": string(data), "update_at": postChanged.UpdateAt,
				"assets": assets})

			return nil
		})
	}
}

func NewPostChaincodeMiddleware(logger *zap.Logger, net common.Network, ipfs *ipfs.IPFSManager, db *gorm.DB) *ChaincodeMiddleware {
	return NewChaincodeMiddleware(logger, net, net.GetContract("post"),

		WithChaincodeHandler("create", "CreatePost", invokeCreatePost(logger, ipfs, db), createPostCallback(logger, ipfs, db)),
		WithChaincodeHandler("update", "UpdatePost", invokeUpdatePost(logger, ipfs, db), updatePostCallback(logger, ipfs, db)),
	)
}
