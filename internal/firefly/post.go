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

type PostRequest struct {
	Content  string   `json:"content"`
	Images   []string `json:"images"`
	ReplyTo  string   `json:"replyTo"`
	BelongTo string   `json:"belongTo"`
}

type FireflyPost struct {
	PostId   string `json:"postId"`
	Content  string `json:"content,omitempty"`
	Operator string `json:"operator,omitempty"`
	BelongTo string `json:"belongTo,omitempty"`
	ReplyTo  string `json:"replyTo,omitempty"`
	Images   string `json:"images,omitempty"`
}

func (ff *FireflyDialer) invokePost(action string, post *FireflyPost) (*resty.Response, proto.MiddlewareError) {
	return ff.invoke(ff.apipoints["Post"], action, post, "")
}

func (ff *FireflyDialer) queryPost(action string, post *FireflyPost) (*resty.Response, proto.MiddlewareError) {
	return ff.query(ff.apipoints["Post"], action, post, "")
}

func (ff *FireflyDialer) CreatePost(c echo.Context) error {

	s, _ := session.Get("session", c)
	userId := s.Values["userId"].(string)

	var postRequest PostRequest

	if c.Bind(&postRequest) != nil {
		return c.JSON(jsonBindingError.Status(), jsonBindingError.Message())
	}

	cid, err := ff.ipfs.Put(strings.NewReader(postRequest.Content))

	if err != nil {
		return c.JSON(err.Status(), err.Message())
	}

	imageSlice, err := ff.uploadImagesBase64ToIPFS(postRequest.Images)

	if err != nil {
		return c.JSON(err.Status(), err.Message())
	}

	images := strings.Join(imageSlice, "-")

	postId := "0x" + hex.EncodeToString(sha256.New().Sum([]byte(postRequest.Content)))

	ffpost := FireflyPost{
		PostId:   postId,
		Content:  cid,
		Operator: userId,
		BelongTo: postRequest.BelongTo,
		ReplyTo:  postRequest.ReplyTo,
		Images:   images,
	}

	resp, err := ff.invokePost("CreatePost", &ffpost)

	if err != nil || resp.StatusCode() != http.StatusAccepted {
		return c.JSON(fireflyInternalError.Status(), fireflyInternalError.Message())
	}

	if err := ff.renewTopicUpdateTime(ffpost.BelongTo); err != nil {
		return c.JSON(err.Status(), err.Message())
	}

	return c.JSON(success.Status(), success.Message())
}

func (ff *FireflyDialer) QueryPostsByBelongTo(c echo.Context) error {
	belongTo := c.QueryParam("belongTo")
	resp, err := ff.queryPost("QueryPostsByBelongTo", &FireflyPost{BelongTo: belongTo})
	if err != nil {
		return c.JSON(err.Status(), err.Message())
	}
	return c.Stream(resp.StatusCode(), "application/json", resp.RawBody())
}
