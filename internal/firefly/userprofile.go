package firefly

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/Cealgull/Middleware/internal/models"
	"github.com/Cealgull/Middleware/internal/proto"
	"github.com/go-resty/resty/v2"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

type UserProfile models.UserProfile

var loginFailureError *LoginFailureError = &LoginFailureError{}

func (ff *FireflyDialer) invokeUserProfile(action string, request *UserProfile) (*resty.Response, proto.MiddlewareError) {
	return ff.invoke(ff.apipoints["UserProfile"], action, request, "")
}

func (ff *FireflyDialer) queryUserProfile(action string, request *UserProfile) (*resty.Response, proto.MiddlewareError) {
	return ff.query(ff.apipoints["UserProfile"], action, request, "")
}

func (ff *FireflyDialer) userRegister(c echo.Context) (*resty.Response, proto.MiddlewareError) {

	s, _ := session.Get("session", c)

	profile := UserProfile{
		UserId:    s.Values["userId"].(string),
		Username:  fmt.Sprintf("User%06d", rand.Intn(1000000)),
		Avatar:    "null",
		Signature: "null",
	}

	resp, err := ff.invokeUserProfile("CreateUser", &profile)
	return resp, err
}

func (ff *FireflyDialer) UserRegister(c echo.Context) error {
	_, err := ff.userRegister(c)
	if err != nil {
		return c.JSON(err.Status(), err.Message())
	}
	return c.JSON(success.Status(), success.Message())
}

func (ff *FireflyDialer) Login(c echo.Context) error {

	s, _ := session.Get("session", c)
	userId := s.Values["userId"].(string)

	resp, err := ff.queryUserProfile("ReadUser", &UserProfile{UserId: userId})

	if err != nil {
		return c.JSON(err.Status(), err.Message())
	}

	if resp.StatusCode() != 200 {
		return c.Stream(http.StatusOK, "application/json", resp.RawBody())
	}

	resp, err = ff.userRegister(c)

	if err != nil {
		return c.JSON(err.Status(), err.Message())
	}

	if resp.StatusCode() == http.StatusAccepted {
		count := 0
		for count < 30 {
			count++
			// wait for the user to be registered
			time.Sleep(100 * time.Millisecond)
			resp, err = ff.queryUserProfile("ReadUser", &UserProfile{UserId: userId})
			if err != nil {
				return c.JSON(err.Status(), err.Message())
			}
			if resp.StatusCode() == 200 {
				return c.Stream(http.StatusOK, "application/json", resp.RawBody())
			}
		}
	}

	return c.JSON(loginFailureError.Status(), loginFailureError.Message())
}

func (ff *FireflyDialer) ReadUserProfile(c echo.Context) error {

	userId := c.QueryParam("userId")

	if userId == "" {
		return c.JSON(jsonBindingError.Status(), jsonBindingError.Message())
	}

	resp, _ := ff.queryUserProfile("ReadUser", &UserProfile{UserId: userId})

	return c.Stream(success.Status(), "application/json", resp.RawBody())

}
