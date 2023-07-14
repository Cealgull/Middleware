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

func topicBaseURL() string {
	return Config.Firefly.Url[0] + Config.Firefly.ApiPrefix + Config.Firefly.ApiName.Topic
}

type CreateTopicReq struct {
	Content  string   `json:"content"`
	Images   []string `json:"images"`
	Title    string   `json:"title"`
	Category string   `json:"category"`
	Tags     []string `json:"tags"`
}

func CreateTopic(c echo.Context) error {
	fmt.Println("CreateTopic Endpoint Hit")

	sess, _ := session.Get("session", c)
	userId := sess.Values["userId"]
	topicId := uuid.New().String()
	topic := CreateTopicReq{}
	err := json.NewDecoder(c.Request().Body).Decode(&topic)
	if err != nil {
		return c.String(http.StatusBadRequest, "Parse JSON error")
	}

	cid, err := ipfs.PutString(topic.Content)
	if err != nil {
		return c.String(http.StatusInternalServerError, "PutString error "+err.Error())
	}

	imageSlice := []string{}
	for _, encodedImage := range topic.Images {
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
	res, err := createTopicImpl(topicId, cid, userId.(string), topic.Title, topic.Category, strings.Join(topic.Tags, "-"), images)
	if err != nil {
		return c.String(http.StatusInternalServerError, "CreateTopic error")
	}
	defer res.Body.Close()

	return c.NoContent(http.StatusNoContent)
}

func createTopicImpl(topicId string, cid string, operator string, title string, category string, tags string, images string) (*http.Response, error) {
	fmt.Println("createTopicImpl Endpoint Hit")
	requestURL := topicBaseURL() + "/invoke/CreateTopic"
	requestBody := fmt.Sprintf(`{"input":{"topicId": "%s", "cid": "%s", "operator": "%s", "title": "%s", "category": "%s", "tags": "%s", "images": "%s"},"key":""}`, topicId, cid, operator, title, category, tags, images)
	return http.Post(requestURL, "application/json", bytes.NewBuffer([]byte(requestBody)))
}

func GetAllTopics(c echo.Context) error {
	fmt.Println("GetAllTopics Endpoint Hit")
	requestURL := topicBaseURL() + "/query/GetAllTopics"
	requestBody := `{"input":{},"key":""}`
	res, err := http.Post(requestURL, "application/json", bytes.NewBuffer([]byte(requestBody)))
	if err != nil {
		return c.String(http.StatusInternalServerError, "GetAllTopics error")
	}
	defer res.Body.Close()

	return c.Stream(res.StatusCode, "application/json", res.Body)
}

func QueryTopicsByTag(c echo.Context) error {
	fmt.Println("QueryTopicsByTag Endpoint Hit")

	tag := c.QueryParam("tag")
	requestURL := topicBaseURL() + "/query/QueryTopicsByTag"
	requestBody := fmt.Sprintf(`{"input":{"tag": "%s"},"key":""}`, tag)
	res, err := http.Post(requestURL, "application/json", bytes.NewBuffer([]byte(requestBody)))
	if err != nil {
		return c.String(http.StatusInternalServerError, "QueryTopicsByTag error")
	}
	defer res.Body.Close()

	return c.Stream(res.StatusCode, "application/json", res.Body)
}
