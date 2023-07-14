package firefly

import (
	"Cealgull_middleware/ipfs"
	"strings"

	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

func postBaseURL() string {
	return Config.Firefly.Url[0] + Config.Firefly.ApiPrefix + Config.Firefly.ApiName.Post
}

type CreatePostReq struct {
	Content  string   `json:"content"`
	Images   []string `json:"images"`
	ReplyTo  string   `json:"replyTo"`
	BelongTo string   `json:"belongTo"`
}

func CreatePost(c echo.Context) error {
	fmt.Println("CreatePost Endpoint Hit")

	sess, _ := session.Get("session", c)
	userId := sess.Values["userId"]
	postId := uuid.New().String()
	post := CreatePostReq{}
	err := json.NewDecoder(c.Request().Body).Decode(&post)
	if err != nil {
		return c.String(http.StatusBadRequest, "Parse JSON error")
	}

	cid, err := ipfs.PutString(post.Content)
	if err != nil {
		return c.String(http.StatusInternalServerError, "PutString error "+err.Error())
	}

	imageSlice := []string{}
	for _, encodedImage := range post.Images {
		decodedImage, err := base64.StdEncoding.DecodeString(encodedImage)
		if err != nil {
			return c.String(http.StatusBadRequest, "base64 DecodeString error "+err.Error())
		}

		buffer := bytes.NewBuffer(decodedImage)
		imageCid, err := ipfs.PutFile(buffer)
		if err != nil {
			return c.String(http.StatusInternalServerError, "PutFile error "+err.Error())
		}
		imageSlice = append(imageSlice, imageCid)
	}
	images := strings.Join(imageSlice, "-")
	res, err := createPostImpl(postId, cid, userId.(string), post.BelongTo, post.ReplyTo, images)
	if err != nil {
		return c.String(http.StatusInternalServerError, "CreatePost error")
	}
	defer res.Body.Close()

	err = RenewTopicUpdateTime(post.BelongTo)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	return c.NoContent(http.StatusNoContent)
}

func createPostImpl(postId string, cid string, operator string, belongTo string, replyTo string, images string) (*http.Response, error) {
	fmt.Println("createPostImpl Endpoint Hit")
	requestURL := postBaseURL() + "/invoke/CreatePost"
	requestBody := fmt.Sprintf(`{"input":{"postId": "%s", "cid": "%s", "operator": "%s", "belongTo": "%s", "replyTo": "%s", "images": "%s"},"key":""}`, postId, cid, operator, belongTo, replyTo, images)
	return http.Post(requestURL, "application/json", bytes.NewBuffer([]byte(requestBody)))
}

func QueryPostsByBelongTo(c echo.Context) error {
	fmt.Println("QueryPostsByBelongTo Endpoint Hit")

	belongTo := c.QueryParam("belongTo")
	requestURL := postBaseURL() + "/query/QueryPostsByBelongTo"
	requestBody := fmt.Sprintf(`{"input":{"belongTo": "%s"},"key":""}`, belongTo)
	res, err := http.Post(requestURL, "application/json", bytes.NewBuffer([]byte(requestBody)))
	if err != nil {
		return c.String(http.StatusInternalServerError, "QueryPostsByBelongTo error")
	}
	defer res.Body.Close()

	return c.Stream(res.StatusCode, "application/json", res.Body)
}

func RenewTopicUpdateTime(topicId string) error {
	fmt.Println("RenewTopicUpdateTime Endpoint Hit")

	requestURL := topicBaseURL() + "/invoke/RenewTopicUpdateTime"
	requestBody := fmt.Sprintf(`{"input":{"topicId": "%s"},"key":""}`, topicId)
	res, err := http.Post(requestURL, "application/json", bytes.NewBuffer([]byte(requestBody)))
	if err != nil {
		return fmt.Errorf("RenewTopicUpdateTime error")
	}
	defer res.Body.Close()

	return nil
}
