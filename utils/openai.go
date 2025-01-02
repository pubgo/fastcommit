package utils

import "github.com/sashabaranov/go-openai"

type OpenaiClient struct {
	Client *openai.Client
	Cfg    *OpenaiConfig
}

type OpenaiConfig struct {
	ApiKey  string `yaml:"api_key"`
	BaseURL string `yaml:"base_url"`
	Model   string `yaml:"model"`
}

func NewOpenaiClient(cfg *OpenaiConfig) *OpenaiClient {
	var openaiCfg = openai.DefaultConfig(cfg.ApiKey)
	openaiCfg.BaseURL = cfg.BaseURL
	return &OpenaiClient{
		Client: openai.NewClientWithConfig(openaiCfg),
		Cfg:    cfg,
	}
}
