package api

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

type ListModelsV1Permission struct {
	ID                 string `json:"id,omitempty"`
	Object             string `json:"object,omitempty"`
	Created            int    `json:"created,omitempty"`
	AllowCreateEngine  bool   `json:"allow_create_engine,omitempty"`
	AllowSampling      bool   `json:"allow_sampling,omitempty"`
	AllowLogprobs      bool   `json:"allow_logprobs,omitempty"`
	AllowSearchIndices bool   `json:"allow_search_indices,omitempty"`
	AllowView          bool   `json:"allow_view,omitempty"`
	AllowFineTuning    bool   `json:"allow_fine_tuning,omitempty"`
	Organization       string `json:"organization,omitempty"`
	Group              any    `json:"group,omitempty"`
	IsBlocking         bool   `json:"is_blocking,omitempty"`
}

type ListModelsV1Data struct {
	ID         string                   `json:"id,omitempty"`
	Object     string                   `json:"object,omitempty"`
	Created    int                      `json:"created,omitempty"`
	OwnedBy    string                   `json:"owned_by,omitempty"`
	Permission []ListModelsV1Permission `json:"permission,omitempty"`
	Root       string                   `json:"root,omitempty"`
	Parent     any                      `json:"parent,omitempty"`
}

type ListModelsV1Output struct {
	Error  *Error             `json:"error,omitempty"`
	Object string             `json:"object,omitempty"`
	Data   []ListModelsV1Data `json:"data,omitempty"`
}

func (impl *ListModelsV1Output) String() string {
	buf := new(bytes.Buffer)
	json.NewEncoder(buf).Encode(impl)
	return buf.String()
}

type ListModelsV1Input struct{}

func (api *OpenAIAPI) ListModelsV1(*ListModelsV1Input) (*ListModelsV1Output, error) {
	endpoint, err := api.endpoint()
	if err != nil {
		return nil, err
	}
	endpoint.Path = "/v1/models"
	req, err := http.NewRequest(
		http.MethodGet,
		endpoint.String(),
		nil,
	)
	if err != nil {
		return nil, err
	}
	api.setToken(req)
	resp, err := api.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		ret := &ListModelsV1Output{}
		if err := json.NewDecoder(resp.Body).Decode(ret); err != nil {
			return nil, err
		}
		return ret, nil
	case http.StatusUnauthorized:
		ret := &ListModelsV1Output{}
		if err := json.NewDecoder(resp.Body).Decode(ret); err != nil {
			return nil, err
		}
		return ret, ErrUnauthorized
	default:
		buf := new(bytes.Buffer)
		io.Copy(buf, resp.Body)
		ret := &ListModelsV1Output{
			Error: &Error{
				Message: buf.String(),
			},
		}
		return ret, ErrUnknown
	}
}
