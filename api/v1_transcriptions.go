package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

func writeField[T any](fieldName string, w *multipart.Writer, v *T) error {
	if v == nil {
		return nil
	}

	return w.WriteField(fieldName, fmt.Sprint(*v))
}

type AudioTranscriptionsV1Input struct {
	File           *os.File
	Model          *string
	Language       *string
	Temperature    *float32
	ResponseFormat *string
	Prompt         *string
}

func (impl *AudioTranscriptionsV1Input) validate() error {
	if impl.File == nil {
		return errors.New("no file")
	} else if impl.Model == nil {
		return errors.New("no models")
	}
	return nil
}

type AudioTranscriptionsV1Output struct {
	Text  string `json:"text,omitempty"`
	Error *Error `json:"error,omitempty"`
}

func (api *OpenAIAPI) AudioTranscriptionsV1(input *AudioTranscriptionsV1Input) (*AudioTranscriptionsV1Output, error) {
	if err := input.validate(); err != nil {
		return nil, err
	}
	defer input.File.Close()
	payload := new(bytes.Buffer)
	writer := multipart.NewWriter(payload)

	part, err := writer.CreateFormFile("file", filepath.Base(input.File.Name()))
	if err != nil {
		return nil, err
	}
	if _, err := io.Copy(part, input.File); err != nil {
		return nil, err
	}
	if err := writer.WriteField("model", *input.Model); err != nil {
		return nil, err
	}
	if err := writeField("language", writer, input.Language); err != nil {
		return nil, err
	}
	if err := writeField("temperature", writer, input.Temperature); err != nil {
		return nil, err
	}
	if err := writeField("response_format", writer, input.ResponseFormat); err != nil {
		return nil, err
	}
	if err := writeField("prompt", writer, input.Prompt); err != nil {
		return nil, err
	}
	if err := writer.Close(); err != nil {
		return nil, err
	}
	endpoint, err := api.endpoint()
	if err != nil {
		return nil, err
	}
	endpoint.Path = "/v1/audio/transcriptions"
	req, err := http.NewRequest(
		http.MethodPost,
		endpoint.String(),
		payload,
	)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	if err := api.setToken(req); err != nil {
		return nil, err
	}
	resp, err := api.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusOK:
		ret := &AudioTranscriptionsV1Output{}
		if err := json.NewDecoder(resp.Body).Decode(ret); err != nil {
			return nil, err
		}
		return nil, err
	case http.StatusUnauthorized:
		ret := &AudioTranscriptionsV1Output{}
		if err := json.NewDecoder(resp.Body).Decode(ret); err != nil {
			return nil, err
		}
		return ret, ErrUnauthorized
	default:
		buf := new(bytes.Buffer)
		io.Copy(buf, resp.Body)
		ret := &AudioTranscriptionsV1Output{
			Error: &Error{
				Message: buf.String(),
			},
		}
		return ret, ErrUnknown
	}
}
