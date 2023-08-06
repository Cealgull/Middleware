package chaincodes

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"

	"github.com/Cealgull/Middleware/internal/ipfs"
	"github.com/Cealgull/Middleware/internal/models"
	"github.com/Cealgull/Middleware/internal/proto"
	"github.com/Cealgull/Middleware/internal/utils"
	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func validateTags(db *gorm.DB, Tags []uint) proto.MiddlewareError {
	tags := []models.TopicTag{}
	if err := db.Find(&tags, Tags).Error; err != nil {
		chaincodeFieldValidationError := ChaincodeFieldValidationFailure{"Tags"}
		return &chaincodeFieldValidationError
	}
	return nil
}

func invokeCreateTopic(logger *zap.Logger, ipfs *ipfs.IPFSManager, db *gorm.DB) ChaincodeInvoke {

	return func(contract *client.Contract, c echo.Context) error {

		type TopicRequest struct {
			Content  string   `json:"content"`
			Images   []string `json:"images"`
			Title    string   `json:"title"`
			Category string   `json:"category"`
			Tags     []uint   `json:"tags"`
		}

		topicRequest := TopicRequest{}

		s, _ := session.Get("session", c)
		wallet := s.Values["wallet"].(string)

		if err := c.Bind(&topicRequest); err != nil {
			return c.JSON(chaincodeDeserializationError.Status(), chaincodeDeserializationError.Message())
		}

		if err := validateTags(db, topicRequest.Tags); err != nil {
			return c.JSON(err.Status(), err.Message())
		}

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

		topicBlock := models.TopicBlock{
			Title:    topicRequest.Title,
			CID:      CID,
			Hash:     "0x" + hex.EncodeToString(sha256.New().Sum([]byte(topicRequest.Content))),
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

		topicBlock := models.TopicBlock{}

		var _ = json.Unmarshal(payload, &topicBlock)

		return db.Transaction(func(tx *gorm.DB) error {

			assets := utils.Map(topicBlock.Images, func(image string) *models.Asset {
				return &models.Asset{
					CID:         image,
					ContentType: "image/jpeg",
				}
			})

			tags := utils.Map(topicBlock.Tags, func(t uint) *models.TopicTag {
				return &models.TopicTag{
					ID:      t,
					Creator: &models.User{Wallet: topicBlock.Creator},
				}
			})

			data, err := ipfs.Cat(topicBlock.CID)

			if err != nil {
				return err
			}

			topic := models.Topic{
				Hash:     topicBlock.Hash,
				Title:    topicBlock.Title,
				Content:  string(data),
				Creator:  &models.User{Wallet: topicBlock.Creator},
				Category: topicBlock.Category,
				Tags:     tags,
				Assets:   assets,
			}

			if err := tx.Create(&topic).Error; err != nil {
				return err
			}

			return nil
		})
	}
}

func NewTopicChaincodeMiddleware(logger *zap.Logger, net *client.Network, ipfs *ipfs.IPFSManager, db *gorm.DB) *ChaincodeMiddleware {
	return NewChaincodeMiddleware(logger, net, "topic",
		WithChaincodeHandler("create", "CreateUser", invokeCreateTopic(logger, ipfs, db), createTopicCallback(logger, ipfs, db)))
}
