package config

var myopenai Openai

type Openai struct {
	ApiKey      string `mapstructure:"apikey"`
	Model       string `mapstructure:"model"`
	ChatPrompt  string `mapstructure:"chat_prompt"`
	Prompt      string `mapstructure:"prompt"`
	ParsePrompt string `mapstructure:"parse_prompt"`
}

func GetOpenaiConf() Openai {
	return myopenai
}
