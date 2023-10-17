package api

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"golang.org/x/xerrors"
)

type ListFileV1Input struct{}

type ListFileV1Data struct {
	ID        string `json:"id,omitempty"`
	Object    string `json:"object,omitempty"`
	Bytes     int    `json:"bytes,omitempty"`
	CreatedAt int    `json:"created_at,omitempty"`
	Filename  string `json:"filename,omitempty"`
	Purpose   string `json:"purpose,omitempty"`
}
type ListFileV1Output struct {
	Data   []ListFileV1Data `json:"data,omitempty"`
	Object *string          `json:"object,omitempty"`
	Error  *Error           `json:"error,omitempty"`
}

func (api *OpenAIAPI) ListFileV1(input *ListFileV1Input) (*ListFileV1Output, error) {
	endpoint, err := api.endpoint()
	if err != nil {
		return nil, err
	}
	endpoint.Path = "/v1/files"
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
		ret := &ListFileV1Output{}
		if err := json.NewDecoder(resp.Body).Decode(ret); err != nil {
			return nil, err
		}
		return ret, nil
	case http.StatusUnauthorized:
		ret := &ListFileV1Output{}
		if err := json.NewDecoder(resp.Body).Decode(ret); err != nil {
			return nil, err
		}
		return ret, ErrUnauthorized
	case http.StatusBadGateway:
		buf := new(bytes.Buffer)
		io.Copy(buf, resp.Body)
		ret := &ListFileV1Output{
			Error: &Error{
				Message: buf.String(),
			},
		}
		return ret, xerrors.Errorf("msg: %s, error: %w", buf.String(), ErrStatusBadGateway)
	default:
		buf := new(bytes.Buffer)
		io.Copy(buf, resp.Body)

		ret := &ListFileV1Output{
			Error: &Error{
				Message: buf.String(),
			},
		}

		return ret, xerrors.Errorf("status_code: %d, msg: %s, error: %w", resp.StatusCode, buf.String(), ErrUnknown)
	}
}
