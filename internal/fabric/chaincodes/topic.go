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

func invokeCreateTopic(logger *zap.Logger, ipfs *ipfs.IPFSManager, db *gorm.DB) ChaincodeInvoke {

	return func(contract common.Contract, c echo.Context) error {

		type TopicRequest struct {
			Content  string   `json:"content"`
			Images   []string `json:"images"`
			Title    string   `json:"title"`
			Category uint     `json:"category"`
			Tags     []uint   `json:"tags"`
		}

		topicRequest := TopicRequest{}

		s, _ := session.Get("session", c)
		wallet := s.Values["wallet"].(string)

		if err := c.Bind(&topicRequest); err != nil {
			return c.JSON(chaincodeDeserializationError.Status(), chaincodeDeserializationError.Message())
		}

		if err := validate(db, []Tag{}, topicRequest.Tags); err != nil {
			return c.JSON(err.Status(), err.Message())
		}

		ts := []byte(time.Now().String())
		ts = append(ts, []byte(wallet)...)

		hash := base64.StdEncoding.EncodeToString(sha256.New().Sum(append([]byte(topicRequest.Content), ts...)))

		CID, err := ipfs.Put(bytes.NewReader([]byte(topicRequest.Content)))

		if err != nil {
			return c.JSON(err.Status(), err.Message())
		}

		images := make([]string, len(topicRequest.Images))

		for i, imageb64 := range topicRequest.Images {

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

		topicBlock := TopicBlock{
			Title:    topicRequest.Title,
			CID:      CID,
			Hash:     hash,
			Creator:  wallet,
			Category: topicRequest.Category,
			Tags:     topicRequest.Tags,
			Images:   images,
		}

		b, _ := json.Marshal(&topicBlock)

		if _, err := contract.Submit("CreateTopic", client.WithBytesArguments(b)); err != nil {
			chaincodeInvokeFailure := ChaincodeInvokeFailureError{"CreateTopic"}
			return c.JSON(chaincodeInvokeFailure.Status(), chaincodeInvokeFailure.Message())
		}

		return c.JSON(success.Status(), success.Message())
	}
}

func createTopicCallback(logger *zap.Logger, ipfs *ipfs.IPFSManager, db *gorm.DB) ChaincodeEventCallback {

	return func(payload []byte) error {

		topicBlock := TopicBlock{}

		var _ = json.Unmarshal(payload, &topicBlock)

		assets := utils.Map(topicBlock.Images, func(image string) *Asset {
			return &Asset{
				CreatorWallet: topicBlock.Creator,
				CID:           image,
				ContentType:   "image/jpeg",
			}
		})

		tagsAssigned := utils.Map(topicBlock.Tags, func(t uint) *TagRelation {
			return &TagRelation{
				TagID: t,
			}
		})

		data, err := ipfs.Cat(topicBlock.CID)

		if err != nil {
			return err
		}

		var _ = assets

		return db.Transaction(func(tx *gorm.DB) error {

			topic := Topic{
				Hash:             topicBlock.Hash,
				Title:            topicBlock.Title,
				Content:          string(data),
				CreatorWallet:    topicBlock.Creator,
				CategoryAssigned: &CategoryRelation{CategoryID: topicBlock.Category},
				TagsAssigned:     tagsAssigned,
				Assets:           assets,
			}

			if err := tx.Create(&topic).Error; err != nil {
				return err
			}

			return nil
		})
	}
}

func NewTopicChaincodeMiddleware(logger *zap.Logger, net common.Network, ipfs *ipfs.IPFSManager, db *gorm.DB) *ChaincodeMiddleware {
	return NewChaincodeMiddleware(logger, net, net.GetContract("topic"),

		WithChaincodeHandler("create", "CreateTopic", invokeCreateTopic(logger, ipfs, db), createTopicCallback(logger, ipfs, db)),
	)
}