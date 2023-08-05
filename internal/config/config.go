package config

type FireflyConfig struct {
	Urls      []string `yaml:"urls"`
	ApiPrefix string   `yaml:"apiPrefix"`
	ApiName   struct {
		Userprofile string `yaml:"userprofile"`
		Topic       string `yaml:"topic"`
		Post        string `yaml:"post"`
	} `yaml:"apiName"`
}

type GatewayConfig struct {
	MspID        string `yaml:"mspID"`
	Channel      string `yaml:"channel"`
	User         string `yaml:"user"`
	CryptoPath   string `yaml:"cryptoPath"`
	PeerEndpoint string `yaml:"peerEndpoint"`
	GatewayPeer  string `yaml:"gatewayPeer"`
}

type PostgresConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
	User string `yaml:"user"`
	Name string `yaml:"name"`
}

type MiddlewareConfig struct {
	Host string `yaml:"string"`
	Port int    `yaml:"port"`
	Ipfs struct {
		URL string `yaml:"url"`
	} `yaml:"ipfs"`
	Firefly  FireflyConfig  `yaml:"firefly"`
	Postgres PostgresConfig `yaml:"postgres"`
	Gateway  GatewayConfig  `yaml:"gateway"`
	Verify   struct {
		URL string `yaml:"url"`
	} `yaml:"verify"`
}
