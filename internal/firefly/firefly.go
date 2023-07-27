package firefly

import (
	"reflect"

	"github.com/Cealgull/Middleware/internal/config"
	"github.com/Cealgull/Middleware/internal/ipfs"
	"github.com/Cealgull/Middleware/internal/proto"
	"github.com/go-resty/resty/v2"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type FireflyDialer struct {
	ipfs      *ipfs.IPFSManager
	logger    *zap.Logger
	client    *resty.Client
	apipoints map[string]string
}

type fireflyPayload struct {
	Input interface{} `json:"input"`
	Key   string      `json:"key"`
}

func NewFireflyDialer(ipfs *ipfs.IPFSManager, logger *zap.Logger, ffconfig *config.FireflyConfig) (*FireflyDialer, error) {

	v := reflect.ValueOf(ffconfig.ApiName)
	apipoints := make(map[string]string)

	for i := 0; i < v.NumField(); i++ {
		apipoints[v.Type().Field(i).Name] = ffconfig.Urls[0] + ffconfig.ApiPrefix + v.Field(i).String()
	}

	ff := &FireflyDialer{
		ipfs:      ipfs,
		client:    resty.New(),
		apipoints: apipoints,
	}

	return ff, nil
}

func (ff *FireflyDialer) invoke(endpoint string, action string, input interface{}, key string) (*resty.Response, proto.MiddlewareError) {
	resp, err := ff.client.R().SetBody(fireflyPayload{Input: input, Key: ""}).Post(endpoint + "/invoke/" + action)
	if err != nil {
		return nil, nil
	}
	if !resp.IsSuccess() {
		return nil, nil
	}
	return resp, nil
}

func (ff *FireflyDialer) query(endpoint string, action string, input interface{}, key string) (*resty.Response, proto.MiddlewareError) {
	resp, err := ff.
		client.
		R().
		SetBody(fireflyPayload{Input: input, Key: ""}).
		Post(endpoint + "/query/" + action)
	if err != nil {
		return nil, nil
	}
	if !resp.IsSuccess() {
		return nil, nil
	}
	return resp, nil
}

func (ff *FireflyDialer) Register(echo *echo.Echo) error {
	echo.POST("/auth/login", ff.Login)

	g := echo.Group("/api")
	list := g.Group("/list")
	query := g.Group("/query")
	create := g.Group("/create")

	list.GET("/topics", ff.GetAllTopics)
	query.GET("/profile", ff.ReadUserProfile)
	query.POST("/queryPostsByBelongTo", ff.QueryPostsByBelongTo)
	query.POST("/queryTopicsByTag", ff.QueryTopicsByTag)
	create.POST("/createTopic", ff.CreateTopic)
	create.POST("/createPost", ff.CreatePost)
	return nil
}
