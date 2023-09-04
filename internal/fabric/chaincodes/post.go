package chaincodes

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"strconv"
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

		if _, err := strconv.Atoi(postRequest.BelongTo); err != nil {
			return c.JSON(chaincodeDeserializationError.Status(), chaincodeDeserializationError.Message())
		}

		if _, err := strconv.Atoi(postRequest.ReplyTo); err != nil {
			return c.JSON(chaincodeDeserializationError.Status(), chaincodeDeserializationError.Message())
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

		belongToID, _ := strconv.Atoi(postBlock.BelongTo)

		return db.Transaction(func(tx *gorm.DB) error {

			post := Post{
				Hash:          postBlock.Hash,
				CreatorWallet: postBlock.Creator,
				Content:       string(data),
				CreateAt:      postBlock.CreateAt,
				UpdateAt:      postBlock.UpdateAt,
				// ReplyTo:
				BelongToID: uint(belongToID),
				Assets:     assets,
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
		type UpdatePostRequest struct {
			Hash    string   `json:"hash"`
			Content string   `json:"content"`
			Images  []string `json:"assets"`
			Type    string   `json:"type"`
		}

		postRequest := UpdatePostRequest{}
		if err := c.Bind(&postRequest); err != nil {
			return c.JSON(chaincodeDeserializationError.Status(), chaincodeDeserializationError.Message())
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
		postBlock := PostBlock{}

		var _ = json.Unmarshal(payload, &postBlock)

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
			post := Post{}
			tx.First(&post, "hash = ?", postBlock.Hash)
			tx.Model(&post).Select("content", "update_at").Updates(map[string]interface{}{"content": string(data), "update_at": postBlock.UpdateAt,
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
