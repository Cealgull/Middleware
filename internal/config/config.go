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

type MiddlewareConfig struct {
	Host string `yaml:"string"`
	Port int    `yaml:"port"`
	Ipfs struct {
		Url string `yaml:"url"`
	} `yaml:"ipfs"`
	Firefly FireflyConfig `yaml:"firefly"`
	Ca      struct {
		Url string `yaml:"url"`
	} `yaml:"ca"`
}
