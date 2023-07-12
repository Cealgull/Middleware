package config

type MiddlewareConfig struct {
	Port string `yaml:"port"`
	Ipfs struct {
		Url string `yaml:"url"`
	} `yaml:"ipfs"`
	Firefly struct {
		Url []string `yaml:"url"`
	} `yaml:"firefly"`
	Ca struct {
		Url string `yaml:"url"`
	} `yaml:"ca"`
}
