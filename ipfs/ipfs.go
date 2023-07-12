package ipfs

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"

	shell "github.com/ipfs/go-ipfs-api"
	"github.com/labstack/echo/v4"
)

var sh *shell.Shell

func Init(url string) {
	fmt.Println("IPFS Init at", url)
	sh = shell.NewShell(url)
}

func Upload(c echo.Context) error {
	fmt.Println("Upload Endpoint Hit")

	inputString := c.FormValue("inputString")
	if inputString != "" {
		return uploadString(c)
	} else {
		return uploadFile(c)
	}
}

func uploadString(c echo.Context) error {
	fmt.Println("Upload String Endpoint Hit")

	inputString := c.FormValue("inputString")
	fmt.Println("UploadString:", inputString)

	cid, err := sh.Add(strings.NewReader(inputString))
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	fmt.Println("CID:", cid)

	return c.String(http.StatusOK, cid)
}

func uploadFile(c echo.Context) error {
	fmt.Println("Upload File Endpoint Hit")

	fileHeader, err := c.FormFile("uploadFile")
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	file, err := fileHeader.Open()
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	defer file.Close()

	fmt.Printf("Uploaded File: %+v\n", fileHeader.Filename)
	fmt.Printf("File Size: %+v\n", fileHeader.Size)
	fmt.Printf("MIME Header: %+v\n", fileHeader.Header)

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	buffer := bytes.NewBuffer(fileBytes)
	cid, err := sh.Add(buffer)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	fmt.Println("CID:", cid)

	return c.String(http.StatusOK, cid)
}
