package ipfs

import (
	"strings"

	"github.com/labstack/echo/v4"
)

const (
	urlEncoded = "application/x-www-urlencoded"
	multipart  = "multipart/form-data"
)

func (m *IPFSManager) upload(c echo.Context) error {

	contentType := c.Request().Header.Get("content-type")

	if len(contentType) != 1 {
		return nil
	}

	if contentType == urlEncoded {

		payload := c.FormValue("payload")
		_, err := m.Put(strings.NewReader(payload))

		if err != nil {
			return c.JSON(err.Status(), err.Message())
		}

	} else if contentType == multipart {

		payload, err := c.FormFile("payload")

		if err != nil {
			return c.JSON(uploadFileMissingError.Status(),
				uploadFileMissingError.Message())
		}

		file, err := payload.Open()

		if err != nil {
			return c.JSON(uploadFileMissingError.Status(),
				uploadFileMissingError.Message())
		}

		if _, err := m.Put(file); err != nil {
			return c.JSON(err.Status(), err.Message())
		}

	} else {
		return c.JSON(uploadHeaderMissingError.Status(), uploadHeaderMissingError.Message())
	}

	return c.JSON(success.Status(), success.Message())
}

func (im *IPFSManager) Register(echo *echo.Echo) error {
	echo.POST("/api/upload", im.upload)
	return nil
}
