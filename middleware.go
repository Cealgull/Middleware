package main

import (
	"Cealgull_middleware/config"
	"Cealgull_middleware/firefly"
	"Cealgull_middleware/ipfs"
	"Cealgull_middleware/verify"

	"fmt"
	"net/http"
	"net/url"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"github.com/spf13/viper"
	"golang.org/x/exp/slices"
)

var e *echo.Echo

var specialEndpoints = []string{
	"/",
	"/upload",
	"/login",
	"/getUserProfile",
	"/createTopic",
	"/createPost",
	"/getAllTopics",
	"/queryPostsByBelongTo",
	"/queryTopicsByTag",
}

var Config config.MiddlewareConfig

func main() {
	// init config
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("/etc/cealgull-middleware")
	viper.AddConfigPath(".")

	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}

	err = viper.Unmarshal(&Config)
	if err != nil {
		panic(err)
	}
	fmt.Println(Config)

	firefly.Config = Config
	verify.Config = Config

	// init ipfs
	ipfs.Init(Config.Ipfs.Url)

	// init firefly targets
	var fireflyTargets []*middleware.ProxyTarget
	for _, fireflyURL := range Config.Firefly.Url {
		target, err := url.Parse(fireflyURL)
		if err != nil {
			panic(err)
		}
		fireflyTargets = append(fireflyTargets, &middleware.ProxyTarget{
			URL: target,
		})
	}

	// init echo
	e = echo.New()
	e.Logger.SetLevel(log.DEBUG)
	e.Use(middleware.Logger())

	e.Use(middleware.ProxyWithConfig(middleware.ProxyConfig{
		Skipper: func(c echo.Context) bool {
			return slices.Contains(specialEndpoints, c.Path())
		},
		Balancer: middleware.NewRoundRobinBalancer(fireflyTargets),
	}))

	e.Use(session.Middleware(sessions.NewCookieStore([]byte("secret"))))
	e.Use(verify.Filter)

	e.GET("/", func(c echo.Context) error {
		verify.InitSession(c, "test")
		return c.String(http.StatusOK, "You are logged in now")
	})
	e.POST("/", func(c echo.Context) error {
		sess, _ := session.Get("session", c)
		ifValid := sess.Values["valid"]

		if ifValid != "valid" {
			return c.String(http.StatusUnauthorized, "Unauthorized")
		}
		return c.String(http.StatusOK, "Hello, World!")
	})
	e.POST("/upload", ipfs.Upload)
	e.POST("/login", verify.Login)
	e.GET("/getUserProfile", firefly.ReadUser)
	e.POST("/createTopic", firefly.CreateTopic)
	e.POST("/createPost", firefly.CreatePost)
	e.GET("/getAllTopics", firefly.GetAllTopics)
	e.GET("/queryPostsByBelongTo", firefly.QueryPostsByBelongTo)
	e.GET("/queryTopicsByTag", firefly.QueryTopicsByTag)

	e.Logger.Fatal(e.Start(fmt.Sprintf(":%s", Config.Port)))
}
