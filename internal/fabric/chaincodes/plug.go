package chaincodes

import (
	"encoding/json"

	"github.com/Cealgull/Middleware/internal/fabric/common"
	"github.com/Cealgull/Middleware/internal/ipfs"
	. "github.com/Cealgull/Middleware/internal/models"
	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func invokeCreateTag(logger *zap.Logger, ipfs *ipfs.IPFSManager, db *gorm.DB) ChaincodeInvoke {

	return func(contract common.Contract, c echo.Context) error {

		type TagRequest struct {
			Name        string `json:"name"`
			Description string `json:"description"`
		}

		tagRequest := TagRequest{}

		if err := c.Bind(&tagRequest); err != nil {
			return c.JSON(chaincodeDeserializationError.Status(), chaincodeDeserializationError.Message())
		}

		s, _ := session.Get("session", c)
		wallet := s.Values["wallet"].(string)

		tagBlock := TagBlock{
			Name:          tagRequest.Name,
			CreatorWallet: wallet,
			Description:   tagRequest.Description,
		}

		b, _ := json.Marshal(&tagBlock)

		if _, err := contract.Submit("CreateTag", client.WithBytesArguments(b)); err != nil {
			chaincodeInvokeFailure := ChaincodeInvokeFailureError{"CreateTag"}
			return c.JSON(chaincodeInvokeFailure.Status(), chaincodeInvokeFailure.Message())
		}

		return c.JSON(success.Status(), success.Message())
	}
}

func createTagCallback(logger *zap.Logger, ipfs *ipfs.IPFSManager, db *gorm.DB) ChaincodeEventCallback {

	return func(payload []byte) error {

		tagBlock := TagBlock{}

		var _ = json.Unmarshal(payload, &tagBlock)

		return db.Transaction(func(tx *gorm.DB) error {

			tag := Tag{
				Name:          tagBlock.Name,
				CreatorWallet: tagBlock.CreatorWallet,
				Description:   tagBlock.Description,
			}

			if err := tx.Create(&tag).Error; err != nil {
				return err
			}

			return nil
		})
	}
}

func NewTagChaincodeMiddleware(logger *zap.Logger, net common.Network, ipfs *ipfs.IPFSManager, db *gorm.DB) *ChaincodeMiddleware {
	return NewChaincodeMiddleware(logger, net, net.GetContract("plug"),

		WithChaincodeHandler("create", "CreateTag", invokeCreateTag(logger, ipfs, db), createTagCallback(logger, ipfs, db)),
	)
}

func invokeCreateCategory(logger *zap.Logger, ipfs *ipfs.IPFSManager, db *gorm.DB) ChaincodeInvoke {

	return func(contract common.Contract, c echo.Context) error {

		type CategoryRequest struct {
			CategoryGroup string `json:"categoryGroup"`
			Color         uint   `json:"color"`
			Name          string `json:"name"`
		}

		categoryRequest := CategoryRequest{}

		if err := c.Bind(&categoryRequest); err != nil {
			return c.JSON(chaincodeDeserializationError.Status(), chaincodeDeserializationError.Message())
		}

		categoryBlock := CategoryBlock{
			CategoryGroupName: categoryRequest.CategoryGroup,
			Color:             categoryRequest.Color,
			Name:              categoryRequest.Name,
		}

		b, _ := json.Marshal(&categoryBlock)

		if _, err := contract.Submit("CreateCategory", client.WithBytesArguments(b)); err != nil {
			chaincodeInvokeFailure := ChaincodeInvokeFailureError{"CreateCategory"}
			return c.JSON(chaincodeInvokeFailure.Status(), chaincodeInvokeFailure.Message())
		}

		return c.JSON(success.Status(), success.Message())
	}
}

func createCategoryCallback(logger *zap.Logger, ipfs *ipfs.IPFSManager, db *gorm.DB) ChaincodeEventCallback {

	return func(payload []byte) error {

		categoryBlock := CategoryBlock{}

		var _ = json.Unmarshal(payload, &categoryBlock)

		return db.Transaction(func(tx *gorm.DB) error {

			category := Category{
				CategoryGroupName: categoryBlock.CategoryGroupName,
				Color:             categoryBlock.Color,
				Name:              categoryBlock.Name,
			}

			if err := tx.Create(&category).Error; err != nil {
				return err
			}

			return nil
		})
	}
}

func NewCategoryChaincodeMiddleware(logger *zap.Logger, net common.Network, ipfs *ipfs.IPFSManager, db *gorm.DB) *ChaincodeMiddleware {
	return NewChaincodeMiddleware(logger, net, net.GetContract("plug"),

		WithChaincodeHandler("create", "CreateCategory", invokeCreateCategory(logger, ipfs, db), createCategoryCallback(logger, ipfs, db)),
	)
}

func invokeCreateCategoryGroup(logger *zap.Logger, ipfs *ipfs.IPFSManager, db *gorm.DB) ChaincodeInvoke {

	return func(contract common.Contract, c echo.Context) error {

		type CategoryGroupRequest struct {
			Name  string `json:"name"`
			Color uint   `json:"color"`
		}

		categoryGroupRequest := CategoryGroupRequest{}

		if err := c.Bind(&categoryGroupRequest); err != nil {
			return c.JSON(chaincodeDeserializationError.Status(), chaincodeDeserializationError.Message())
		}

		categoryGroupBlock := CategoryGroupBlock{
			Name:  categoryGroupRequest.Name,
			Color: categoryGroupRequest.Color,
		}

		b, _ := json.Marshal(&categoryGroupBlock)

		if _, err := contract.Submit("CreateCategoryGroup", client.WithBytesArguments(b)); err != nil {
			chaincodeInvokeFailure := ChaincodeInvokeFailureError{"CreateCategoryGroup"}
			return c.JSON(chaincodeInvokeFailure.Status(), chaincodeInvokeFailure.Message())
		}

		return c.JSON(success.Status(), success.Message())
	}
}

func createCategoryGroupCallback(logger *zap.Logger, ipfs *ipfs.IPFSManager, db *gorm.DB) ChaincodeEventCallback {

	return func(payload []byte) error {

		categoryGroupBlock := CategoryGroupBlock{}

		var _ = json.Unmarshal(payload, &categoryGroupBlock)

		return db.Transaction(func(tx *gorm.DB) error {

			categoryGroup := CategoryGroup{
				Color: categoryGroupBlock.Color,
				Name:  categoryGroupBlock.Name,
			}

			if err := tx.Create(&categoryGroup).Error; err != nil {
				return err
			}

			return nil
		})
	}
}

func NewCategoryGroupChaincodeMiddleware(logger *zap.Logger, net common.Network, ipfs *ipfs.IPFSManager, db *gorm.DB) *ChaincodeMiddleware {
	return NewChaincodeMiddleware(logger, net, net.GetContract("plug"),

		WithChaincodeHandler("create", "CreateCategoryGroup", invokeCreateCategoryGroup(logger, ipfs, db), createCategoryGroupCallback(logger, ipfs, db)),
	)
}
