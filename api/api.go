package api

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/ieee0824/gopenai-api/config"
	"golang.org/x/xerrors"
)

type OpenAIAPI struct {
	httpClient    *http.Client // default: http.DefaultClient
	configuration *config.Configuration
}

func New(cfg *config.Configuration) OpenAIAPIIface {

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
		return xerrors.New("no token")
	}
	if api.configuration.Organization == nil {
		return xerrors.New("no organization")
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", *api.configuration.ApiKey))
	req.Header.Add("OpenAI-Organization", *api.configuration.Organization)
	return nil
}
