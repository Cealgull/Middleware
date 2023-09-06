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
			Category string   `json:"category"`
			Tags     []string `json:"tags"`
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

		type TopicResponse struct {
			Hash string `json:"hash"`
		}

		return c.JSON(success.Status(), &TopicResponse{Hash: hash})
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

		tagsAssigned := utils.Map(topicBlock.Tags, func(t string) *TagRelation {
			return &TagRelation{
				TagName: t,
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
				CategoryAssigned: &CategoryRelation{CategoryName: topicBlock.Category},
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

func invokeUpdateTopic(logger *zap.Logger, ipfs *ipfs.IPFSManager, db *gorm.DB) ChaincodeInvoke {
	return func(contract common.Contract, c echo.Context) error {
		type ChangeTopicRequest struct {
			Hash    string   `json:"hash"`
			Content string   `json:"content"`
			Assets  []string `json:"assets"`
		}

		topicRequest := ChangeTopicRequest{}
		if err := c.Bind(&topicRequest); err != nil {
			return c.JSON(chaincodeDeserializationError.Status(), chaincodeDeserializationError.Message())
		}

		topic := Topic{}
		if err := db.Model(&Topic{}).
			Where("hash = ?", topicRequest.Hash).First(&topic).Error; err != nil {
			return err
		}

		CID, err := ipfs.Put(bytes.NewReader([]byte(topicRequest.Content)))

		if err != nil {
			return c.JSON(err.Status(), err.Message())
		}

		images := make([]string, len(topicRequest.Assets))

		for i, imageb64 := range topicRequest.Assets {

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
			Hash:   topicRequest.Hash,
			CID:    CID,
			Images: images,
		}

		b, _ := json.Marshal(&topicBlock)

		if _, err := contract.Submit("UpdateTopic", client.WithBytesArguments(b)); err != nil {
			chaincodeInvokeFailure := ChaincodeInvokeFailureError{"UpdateTopic"}
			return c.JSON(chaincodeInvokeFailure.Status(), chaincodeInvokeFailure.Message())
		}

		return c.JSON(success.Status(), success.Message())

	}
}

func updateTopicCallback(logger *zap.Logger, ipfs *ipfs.IPFSManager, db *gorm.DB) ChaincodeEventCallback {

	return func(payload []byte) error {
		topicChanged := TopicBlock{}

		var _ = json.Unmarshal(payload, &topicChanged)

		topic := Topic{}
		if err := db.Model(&Topic{}).
			Where("hash = ?", topicChanged.Hash).First(&topic).Error; err != nil {
			return err
		}

		data, err := ipfs.Cat(topicChanged.CID)
		if err != nil {
			return err
		}

		images := utils.Map(topicChanged.Images, func(image string) *Asset {
			return &Asset{
				CreatorWallet: topic.CreatorWallet,
				CID:           image,
				ContentType:   "image/jpeg",
			}
		})

		var _ = images

		return db.Transaction(func(tx *gorm.DB) error {
			tx.Model(&topic).Select("content", "update_at", "images").Updates(map[string]interface{}{"content": string(data), "images": images})

			return nil
		})
	}
}

func invokeUpvoteTopic(logger *zap.Logger, db *gorm.DB) ChaincodeInvoke {
	return func(contract common.Contract, c echo.Context) error {
		type UpvoteRequest struct {
			Hash string `json:"hash"`
		}

		upvoteRequest := UpvoteRequest{}
		if err := c.Bind(&upvoteRequest); err != nil {
			return c.JSON(chaincodeDeserializationError.Status(), chaincodeDeserializationError.Message())
		}

		topic := Topic{}
		if err := db.Model(&Topic{}).
			Where("hash = ?", upvoteRequest.Hash).First(&topic).Error; err != nil {
      return err
		}

		s, _ := session.Get("session", c)
		wallet := s.Values["wallet"].(string)

		upvoteBlock := UpvoteBlock{
			Hash:    upvoteRequest.Hash,
			Creator: wallet,
		}
		b, _ := json.Marshal(&upvoteBlock)

		if _, err := contract.Submit("UpvoteTopic", client.WithBytesArguments(b)); err != nil {
			chaincodeInvokeFailure := ChaincodeInvokeFailureError{"UpvoteTopic"}
			return c.JSON(chaincodeInvokeFailure.Status(), chaincodeInvokeFailure.Message())
		}

		return c.JSON(success.Status(), success.Message())

	}
}

// FIXME: WE NEED TO UNDO DOWNVOTE IF UPVOTED?
func upvoteTopicCallback(logger *zap.Logger, db *gorm.DB) ChaincodeEventCallback {

	return func(payload []byte) error {

		upvoteBlock := UpvoteBlock{}

		var _ = json.Unmarshal(payload, &upvoteBlock)

		topic := Topic{}

		if err := db.Model(&Topic{}).
			Preload("Upvotes").
			Where("hash = ?", upvoteBlock.Hash).First(&topic).Error; err != nil {
			return err
		}

		return db.Transaction(func(tx *gorm.DB) error {

			upvote := Upvote{
				CreatorWallet: upvoteBlock.Creator,
			}

			for _, u := range topic.Upvotes {
				if u.CreatorWallet == upvote.CreatorWallet {
					tx.Model(&topic).Association("Upvotes").Delete(u)
					return nil
				}
			}

			var _ = tx.Model(&topic).Association("Upvotes").Append(&upvote)

			return nil
		})
	}
}

// FIXME: WE NEED TO UNDO UPVOTE IF DOWNVOTED?
func invokeDownvoteTopic(logger *zap.Logger, db *gorm.DB) ChaincodeInvoke {
	return func(contract common.Contract, c echo.Context) error {
		type DownvoteRequest struct {
			Hash string `json:"hash"`
		}

		downvoteRequest := DownvoteRequest{}
		if err := c.Bind(&downvoteRequest); err != nil {
			return c.JSON(chaincodeDeserializationError.Status(), chaincodeDeserializationError.Message())
		}

		topic := Topic{}
		if err := db.Model(&Topic{}).
			Where("hash = ?", downvoteRequest.Hash).First(&topic).Error; err != nil {
			return err
		}

		s, _ := session.Get("session", c)
		wallet := s.Values["wallet"].(string)

		downvoteBlock := DownvoteBlock{
			Hash:    downvoteRequest.Hash,
			Creator: wallet,
		}
		b, _ := json.Marshal(&downvoteBlock)

		if _, err := contract.Submit("DownvoteTopic", client.WithBytesArguments(b)); err != nil {
			chaincodeInvokeFailure := ChaincodeInvokeFailureError{"DownvoteTopic"}
			return c.JSON(chaincodeInvokeFailure.Status(), chaincodeInvokeFailure.Message())
		}

		return c.JSON(success.Status(), success.Message())

	}
}

func downvoteTopicCallback(logger *zap.Logger, db *gorm.DB) ChaincodeEventCallback {

	return func(payload []byte) error {

		downvoteBlock := DownvoteBlock{}

		var _ = json.Unmarshal(payload, &downvoteBlock)

		topic := Topic{}

		if err := db.Model(&Topic{}).
			Preload("Downvotes").
			Where("hash = ?", downvoteBlock.Hash).First(&topic).Error; err != nil {
			return err
		}

		return db.Transaction(func(tx *gorm.DB) error {

			downvote := Downvote{
				CreatorWallet: downvoteBlock.Creator,
			}

			for _, d := range topic.Upvotes {
				if d.CreatorWallet == downvote.CreatorWallet {
					tx.Model(&topic).Association("Downvotes").Delete(d)
					return nil
				}
			}
			var _ = tx.Model(&topic).Association("Downvotes").Append(&downvote)
			return nil
		})
	}
}

func queryCategories(logger *zap.Logger, db *gorm.DB) ChaincodeQuery {
	return func(c echo.Context) error {
		categories := []Category{}
		var _ = db.Find(&categories).Error

		return c.JSON(success.Status(), categories)
	}
}

func queryTags(logger *zap.Logger, db *gorm.DB) ChaincodeQuery {
	return func(c echo.Context) error {
		tags := []Tag{}
		var _ = db.Find(&tags).Error

		return c.JSON(success.Status(), tags)
	}
}

func queryTopicsList(logger *zap.Logger, db *gorm.DB) ChaincodeQuery {
	return func(c echo.Context) error {

		type QueryRequest struct {
			PageOrdinal int      `json:"pageOrdinal"`
			PageSize    int      `json:"pageSize"`
			Category    string   `json:"category"`
			Creator     string   `json:"creator"`
			Tags        []string `json:"tags"`
		}

		q := QueryRequest{}

		if c.Bind(&q) != nil {
      return c.JSON(chaincodeDeserializationError.Status(), chaincodeDeserializationError.Message())
		}
    
    if q.PageOrdinal <= 0 || q.PageSize <= 0{
      return c.JSON(chaincodeQueryParameterError.Status(), chaincodeQueryParameterError.Message())
    }


    topics := []Topic{}

		err := db.Transaction(func(tx *gorm.DB) error {

      tx = tx.Model(&Topic{}).
      Preload("CategoryAssigned").
      Preload("TagsAssigned").
      Preload("Upvotes").
      Preload("Downvotes").
      Scopes(paginate(q.PageOrdinal, q.PageSize))
      
      if q.Creator != "" {
        tx = tx.Where("creator_wallet = ?", q.Creator)
      }

      if q.Category != "" { 
        subquery := db.Select("TopicID").Model(&CategoryRelation{}).Where("category_name = ?", q.Category)
        tx = tx.Joins("inner join (?) as t1 on t1.topic_id = topics.id", subquery)
      }

      if len(q.Tags) != 0 {
        subquery := db.Select("OwnerID").Model(&TagRelation{}).Where("tag_name IN ?", q.Tags)
        tx = tx.InnerJoins("inner join (?) as t2 on t2.owner_id = topics.id", subquery)
      }


      if err := tx.Find(&topics).Error; err != nil {
        return err
      }

      return nil
		})

    if err != nil {
      return c.JSON(chaincodeInternalError.Status(), chaincodeInternalError.Message())
    }

		return c.JSON(success.Status(), topics)

	}
}

func NewTopicChaincodeMiddleware(logger *zap.Logger, net common.Network, ipfs *ipfs.IPFSManager, db *gorm.DB) *ChaincodeMiddleware {
	return NewChaincodeMiddleware(logger, net, net.GetContract("topic"),

		WithChaincodeHandler("create", "CreateTopic", invokeCreateTopic(logger, ipfs, db), createTopicCallback(logger, ipfs, db)),
		WithChaincodeHandler("update", "UpdateTopic", invokeUpdateTopic(logger, ipfs, db), updateTopicCallback(logger, ipfs, db)),

		WithChaincodeHandler("upvote", "UpvoteTopic", invokeUpvoteTopic(logger, db), upvoteTopicCallback(logger, db)),
		WithChaincodeHandler("downvote", "DownvoteTopic", invokeDownvoteTopic(logger, db), downvoteTopicCallback(logger, db)),

		WithChaincodeQuery("categories", queryCategories(logger, db)),
		WithChaincodeQuery("tags", queryTags(logger, db)),
    WithChaincodeQuery("list", queryTopicsList(logger, db)),
	)
}
