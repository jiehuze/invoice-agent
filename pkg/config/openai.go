package config

var myopenai Openai

type Openai struct {
	ApiKey string `mapstructure:"apikey"`
	Model  string `mapstructure:"model"`
	Prompt string `mapstructure:"prompt"`
}

func GetOpenaiConf() Openai {
	return myopenai
}
