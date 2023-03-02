package api

import (
	"net/http"

	"github.com/ieee0824/gopenai-api/config"
)

type OpenAIAPI struct {
	httpClient    *http.Client // default: http.DefaultClient
	configuration *config.Configuration
}

func New(cfg *config.Configuration) *OpenAIAPI {
	return &OpenAIAPI{
		httpClient:    http.DefaultClient,
		configuration: cfg,
	}
}
