package firefly

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"

	"github.com/Cealgull/Middleware/internal/proto"
	"github.com/go-resty/resty/v2"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

type TopicRequest struct {
	Content  string   `json:"content"`
	Images   []string `json:"images"`
	Title    string   `json:"title"`
	Category string   `json:"category"`
	Tags     []string `json:"tags"`
}

type FireflyTopic struct {
	TopicId  string `json:"topicId,omitempty"`
	Cid      string `json:"cid,omitempty"`
	UserId   string `json:"userId,omitempty"`
	Title    string `json:"title,omitempty"`
	Category string `json:"category,omitempty"`
}

type FireflyTopicQuery struct {
	FireflyTopic
	Tag string `json:"tag,omitempty"`
}

type FireflyTopicInvoke struct {
	FireflyTopic
	Tags   string `json:"tags,omitempty"`
	Images string `json:"images,omitempty"`
}

func (ff *FireflyDialer) invokeTopic(action string, topic *FireflyTopicInvoke) (*resty.Response, proto.MiddlewareError) {
	return ff.invoke(ff.apipoints["Topic"], action, topic, "")
}

func (ff *FireflyDialer) queryTopic(action string, topic *FireflyTopicQuery) (*resty.Response, proto.MiddlewareError) {
	return ff.query(ff.apipoints["Topic"], action, topic, "")
}

func (ff *FireflyDialer) CreateTopic(c echo.Context) error {

	ff.logger.Debug("Creating Topic...")

	s, _ := session.Get("session", c)
	userId := s.Values["userId"].(string)

	var topicRequest TopicRequest

	if c.Bind(&topicRequest) != nil {
		return c.JSON(jsonBindingError.Status(), jsonBindingError.Message())
	}

	topicId := "0x" + hex.EncodeToString(sha256.New().Sum([]byte(topicRequest.Content)))

	cid, err := ff.ipfs.Put(strings.NewReader(topicRequest.Content))

	if err != nil {
		return c.JSON(err.Status(), err.Message())
	}

	imageSlice, err := ff.uploadImagesBase64ToIPFS(topicRequest.Images)

	if err != nil {
		return c.JSON(err.Status(), err.Message())
	}

	images := strings.Join(imageSlice, "-")

	ffTopic := FireflyTopicInvoke{
		FireflyTopic: FireflyTopic{TopicId: topicId,
			Cid:      cid,
			UserId:   userId,
			Title:    topicRequest.Title,
			Category: topicRequest.Category},
		Tags:   strings.Join(topicRequest.Tags, "-"),
		Images: images,
	}

	resp, err := ff.invokeTopic("CreateTopic", &ffTopic)

	if err != nil || resp.StatusCode() != http.StatusAccepted {
		return c.JSON(fireflyInternalError.Status(), fireflyInternalError.Message())
	}

	return c.JSON(success.Status(), success.Message())

}

func (ff *FireflyDialer) GetAllTopics(c echo.Context) error {
	resp, err := ff.queryTopic("GetAllTopics", &FireflyTopicQuery{})
	if err != nil {
		return c.JSON(err.Status(), err.Message())
	}
	return c.Stream(resp.StatusCode(), "application/json", resp.RawBody())
}

func (ff *FireflyDialer) QueryTopicsByTag(c echo.Context) error {
	tag := c.QueryParam("tag")
	resp, err := ff.queryTopic("QueryTopicByTag", &FireflyTopicQuery{Tag: tag})
	if err != nil {
		return c.JSON(err.Status(), err.Message())
	}
	return c.Stream(resp.StatusCode(), "application/json", resp.RawBody())
}

func (ff *FireflyDialer) renewTopicUpdateTime(topicId string) proto.MiddlewareError {
	_, err := ff.invokeTopic("RenewTopicUpdateTime", &FireflyTopicInvoke{FireflyTopic: FireflyTopic{TopicId: topicId}})
	return err
}
