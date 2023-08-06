package main

import (
	"github.com/Cealgull/Middleware/internal/authority"
	"github.com/Cealgull/Middleware/internal/config"
	"github.com/Cealgull/Middleware/internal/fabric"
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

	config := config.MiddlewareConfig{}

	logger, _ := zap.NewProduction()

	err := viper.ReadInConfig()

	if err != nil {
		logger.Panic(err.Error())
	}

	err = viper.Unmarshal(&config)

	if err != nil {
		logger.Panic(err.Error())
	}

	ipfs, _ := ipfs.NewIPFSManager(logger, config.Ipfs.URL)
	ca := authority.NewCertAuthority(logger, config.Verify.URL)

	fab, err := fabric.NewGatewayMiddleware(logger, ipfs, &config)

	if err != nil {
		logger.Panic(err.Error())
	}

	r, _ := rest.NewRestServer(config.Host,
		config.Port,
		rest.WithEndpoint(ca),
		rest.WithEndpoint(ipfs),
		rest.WithEndpoint(fab),
	)

	r.Start()

}
