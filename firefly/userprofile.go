package firefly

import (
	"Cealgull_middleware/config"

	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

var Config config.MiddlewareConfig

func userprofileBaseURL() string {
	return Config.Firefly.Url[0] + Config.Firefly.ApiPrefix + Config.Firefly.ApiName.Userprofile
}

func Register(c echo.Context) (*http.Response, error) {
	fmt.Println("Register Endpoint Hit")

	sess, _ := session.Get("session", c)
	userId := sess.Values["userId"]
	// random username by default
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	username := "用户" + strconv.Itoa(r.Intn(900000)+100000)
	avatar := "QmfTT625M15eaWUXPqkvZZZQWPrDvRa68GL2FBk6qD6sJr"
	signature := ""

	requestURL := userprofileBaseURL() + "/invoke/CreateUser"
	requestBody := fmt.Sprintf(`{"input":{"userId": "%s", "username": "%s", "avatar": "%s", "signature": "%s"},"key":""}`, userId, username, avatar, signature)
	return http.Post(requestURL, "application/json", bytes.NewBuffer([]byte(requestBody)))
}

func ReadUser(c echo.Context) error {
	fmt.Println("ReadUser Endpoint Hit")

	jsonMap := make(map[string]interface{})
	err := json.NewDecoder(c.Request().Body).Decode(&jsonMap)
	if err != nil {
		return c.String(http.StatusBadRequest, "Parse JSON error")
	}
	userId := jsonMap["userId"].(string)
	if userId == "" {
		return c.String(http.StatusBadRequest, "userId not found")
	}

	res, err := readUserImpl(userId)
	if err != nil {
		return c.String(http.StatusInternalServerError, "ReadUser error")
	}
	defer res.Body.Close()
	return c.Stream(res.StatusCode, "application/json", res.Body)
}

func ReadUserLogin(c echo.Context, userId string) (*http.Response, error) {
	fmt.Println("ReadUserLogin Endpoint Hit")
	return readUserImpl(userId)
}

func readUserImpl(userId string) (*http.Response, error) {
	fmt.Println("readUserImpl Endpoint Hit")
	requestURL := userprofileBaseURL() + "/query/ReadUser"
	requestBody := fmt.Sprintf(`{"input":{"userId": "%s"},"key":""}`, userId)
	return http.Post(requestURL, "application/json", bytes.NewBuffer([]byte(requestBody)))
}
