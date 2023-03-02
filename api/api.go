package api

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"

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

func (api *OpenAIAPI) endpoint() (*url.URL, error) {
	if api.configuration.Endpoint == nil {
		return url.Parse("https://api.openai.com")
	}
	return url.Parse(*api.configuration.Endpoint)
}

func (api *OpenAIAPI) setToken(req *http.Request) error {
	if api.configuration.ApiKey == nil {
		return errors.New("no token")
	}
	if api.configuration.Organization == nil {
		return errors.New("no organization")
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", *api.configuration.ApiKey))
	req.Header.Add("OpenAI-Organization", *api.configuration.Organization)
	return nil
}
