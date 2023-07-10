package firefly

import (
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strconv"
	"time"

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

func Register(c echo.Context) error {
	fmt.Println("Register Endpoint Hit")

	userId := c.FormValue("userId")
	// random username by default
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	username := "用户" + strconv.Itoa(r.Intn(900000)+100000)
	avatar := "https://github.githubassets.com/images/modules/logos_page/GitHub-Mark.png"
	signature := ""

	requestURL := "http://127.0.0.1:5000/api/v1/namespaces/default/apis/userprofile011/invoke/CreateUser"
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
