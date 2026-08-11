package utils

import (
	"net/http"
	"time"

	"github.com/sashabaranov/go-openai"
)

type OpenaiClient struct {
	Client *openai.Client
	Cfg    *OpenaiConfig
}

type OpenaiConfig struct {
	ApiKey  string `yaml:"api_key"`
	BaseURL string `yaml:"base_url"`
	Model   string `yaml:"model"`
}

const defaultOpenAITimeout = 45 * time.Second

func NewOpenaiClient(cfg *OpenaiConfig) *OpenaiClient {
	var openaiCfg = openai.DefaultConfig(cfg.ApiKey)
	openaiCfg.BaseURL = cfg.BaseURL
	openaiCfg.HTTPClient = &http.Client{Timeout: defaultOpenAITimeout}
	return &OpenaiClient{
		Client: openai.NewClientWithConfig(openaiCfg),
		Cfg:    cfg,
	}
}
