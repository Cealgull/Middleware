package config

type MiddlewareConfig struct {
	Port string `yaml:"port"`
	Ipfs struct {
		Url string `yaml:"url"`
	} `yaml:"ipfs"`
	Firefly struct {
		Url       []string `yaml:"url"`
		ApiPrefix string   `yaml:"apiPrefix"`
		ApiName   struct {
			Userprofile string `yaml:"userprofile"`
			Topic       string `yaml:"topic"`
			Post        string `yaml:"post"`
		} `yaml:"apiName"`
	} `yaml:"firefly"`
	Ca struct {
		Url string `yaml:"url"`
	} `yaml:"ca"`
}
