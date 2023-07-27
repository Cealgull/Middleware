package firefly

import (
	"bytes"
	"encoding/base64"

	"github.com/Cealgull/Middleware/internal/proto"
)

func (ff *FireflyDialer) uploadImagesBase64ToIPFS(images []string) ([]string, proto.MiddlewareError) {

	imageSlice := []string{}

	for _, encodedImage := range images {

		decodedImage, err := base64.StdEncoding.DecodeString(encodedImage)

		if err != nil {
			return nil, &Base64DecodeError{}
		}

		buffer := bytes.NewBuffer(decodedImage)

		if imageCid, err := ff.ipfs.Put(buffer); err != nil {
			return nil, err
		} else {
			imageSlice = append(imageSlice, imageCid)
		}
	}

	return imageSlice, nil
}
