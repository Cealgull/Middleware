package config

type GatewayConfig struct {
	MspID        string `yaml:"mspID"`
	Channel      string `yaml:"channel"`
	User         string `yaml:"user"`
	CryptoPath   string `yaml:"cryptoPath"`
	PeerEndpoint string `yaml:"peerEndpoint"`
	GatewayPeer  string `yaml:"gatewayPeer"`
}

type PostgresDSNConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
	User string `yaml:"user"`
	Name string `yaml:"name"`
	Seed bool   `yaml:"seed"`
}

type IPFSConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

type VerifyConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

type MiddlewareConfig struct {
	Host     string            `yaml:"host"`
	Port     int               `yaml:"port"`
	IPFS     IPFSConfig        `yaml:"ipfs"`
	Postgres PostgresDSNConfig `yaml:"postgres"`
	Gateway  GatewayConfig     `yaml:"gateway"`
	Verify   VerifyConfig      `yaml:"verify"`
}
