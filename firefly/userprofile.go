package firefly

import (
	"Cealgull_middleware/config"

	"bytes"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

type UserProfile struct {
	IdentityId string   `json:"identityId"`
	Username   string   `json:"username"`
	Avatar     string   `json:"avatar"`
	Signature  string   `json:"signature"`
	Roles      []string `json:"roles,omitempty" metadata:"roles,optional" `
	Badge      []string `json:"badge,omitempty" metadata:"badge,optional" `
}

var Config config.MiddlewareConfig

var baseURL = Config.Firefly.Url[0] + Config.Firefly.ApiPrefix + Config.Firefly.ApiName.Userprofile

func Register(c echo.Context) error {
	fmt.Println("Register Endpoint Hit")

	sess, _ := session.Get("session", c)
	userId := sess.Values["userId"]
	// random username by default
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	username := "用户" + strconv.Itoa(r.Intn(900000)+100000)
	avatar := "https://github.githubassets.com/images/modules/logos_page/GitHub-Mark.png"
	signature := ""

	requestURL := baseURL + "/invoke/CreateUser"
	requestBody := fmt.Sprintf(`"input":{"userId": "%s", "username": "%s", "avatar": "%s", "signature": "%s"},"key":""`, userId, username, avatar, signature)
	res, err := http.Post(requestURL, "application/json", bytes.NewBuffer([]byte(requestBody)))
	if err != nil {
		fmt.Println("Error:", err)
		return err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println("Error:", err)
		return err
	}

	fmt.Println("Response Body:", string(body))

	return c.JSON(res.StatusCode, string(body))
}
