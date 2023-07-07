package ipfs

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	shell "github.com/ipfs/go-ipfs-api"
	"github.com/labstack/echo/v4"
)

var sh *shell.Shell

func Init(url string) {
	fmt.Println("IPFS Init")
	sh = shell.NewShell(url)
}

func Upload(c echo.Context) error {
	fmt.Println("Upload Endpoint Hit")

	fileHeader, err := c.FormFile("file")
	if err != nil {
		return err
	}

	file, err := fileHeader.Open()
	if err != nil {
		fmt.Println("Error opening file")
		return err
	}
	defer file.Close()

	fmt.Printf("Uploaded File: %+v\n", fileHeader.Filename)
	fmt.Printf("File Size: %+v\n", fileHeader.Size)
	fmt.Printf("MIME Header: %+v\n", fileHeader.Header)

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		fmt.Println("Error reading file")
		return err
	}

	buffer := bytes.NewBuffer(fileBytes)
	cid, err := sh.Add(buffer)
	if err != nil {
		return err
	}
	fmt.Println("CID:", cid)

	return c.String(http.StatusOK, cid)
}
