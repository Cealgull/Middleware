package main

import (
	"github.com/Cealgull/Middleware/internal/authority"
	"github.com/Cealgull/Middleware/internal/config"
	"github.com/Cealgull/Middleware/internal/firefly"
	"github.com/Cealgull/Middleware/internal/ipfs"
	"github.com/Cealgull/Middleware/internal/rest"
	"go.uber.org/zap"

	"github.com/spf13/viper"
)

func main() {

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("/etc/cealgull-middleware")
	viper.AddConfigPath(".")

	var config config.MiddlewareConfig
	logger, _ := zap.NewProduction()

	err := viper.ReadInConfig()

	if err != nil {
		logger.Panic(err.Error())
	}

	err = viper.Unmarshal(&config)

	if err != nil {
		logger.Panic(err.Error())
	}

	im, _ := ipfs.NewIPFSManager(config.Ipfs.Url, logger)
	ca := authority.NewCertAuthority(logger, config.Ca.Url)
	ff, _ := firefly.NewFireflyDialer(im, logger, &config.Firefly)

	r, _ := rest.NewRestServer(config.Host,
		config.Port,
		rest.WithLogger(logger),
		rest.WithEndpoint(ca),
		rest.WithEndpoint(im),
		rest.WithEndpoint(ff),
	)

	r.Start()

}
