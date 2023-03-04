package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

type ImagesGenerationsV1Input struct {
	Prompt         *string `json:"prompt,omitempty"`
	N              *int    `json:"n,omitempty"`
	Size           *string `json:"size,omitempty"`
	ResponseFormat *string `json:"response_format,omitempty"`
	User           *string `json:"user,omitempty"`
}

func (impl *ImagesGenerationsV1Input) validate() error {
	if impl.Prompt == nil {
		return errors.New("no prompt")
	}
	return nil
}

type ImagesGenerationsV1Output struct {
	Created int `json:"created,omitempty"`
	Data    []struct {
		URL string `json:"url,omitempty"`
	} `json:"data,omitempty"`
	Error *Error `json:"error,omitempty"`
}

func (api *OpenAIAPI) ImagesGenerationsV1(input *ImagesGenerationsV1Input) (*ImagesGenerationsV1Output, error) {
	if err := input.validate(); err != nil {
		return nil, err
	}
	reqBody := new(bytes.Buffer)
	if err := json.NewEncoder(reqBody).Encode(input); err != nil {
		return nil, err
	}
	endpoint, err := api.endpoint()
	if err != nil {
		return nil, err
	}
	endpoint.Path = "/v1/images/generations"
	req, err := http.NewRequest(
		http.MethodPost,
		endpoint.String(),
		reqBody,
	)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if err := api.setToken(req); err != nil {
		return nil, err
	}
	resp, err := api.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	ret := &ImagesGenerationsV1Output{}
	switch resp.StatusCode {
	case http.StatusOK:
		if err := json.NewDecoder(resp.Body).Decode(ret); err != nil {
			return nil, err
		}
		return ret, nil
	case http.StatusUnauthorized:
		if err := json.NewDecoder(resp.Body).Decode(ret); err != nil {
			return nil, err
		}
		return ret, ErrUnauthorized
	default:
		buf := new(bytes.Buffer)
		io.Copy(buf, resp.Body)
		ret.Error = &Error{
			Message: buf.String(),
		}
		return ret, nil
	}

}
