package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"golang.org/x/xerrors"
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
		return xerrors.New("no file")
	} else if impl.Model == nil {
		return xerrors.New("no models")
	}
	return nil
}

type AudioTranscriptionsV1Segments struct {
	ID               int     `json:"id,omitempty"`
	Seek             float32 `json:"seek,omitempty"`
	Start            float32 `json:"start,omitempty"`
	End              float32 `json:"end,omitempty"`
	Text             string  `json:"text,omitempty"`
	Tokens           []int   `json:"tokens,omitempty"`
	Temperature      float32 `json:"temperature,omitempty"`
	AvgLogprob       float64 `json:"avg_logprob,omitempty"`
	CompressionRatio float64 `json:"compression_ratio,omitempty"`
	NoSpeechProb     float64 `json:"no_speech_prob,omitempty"`
	Transient        bool    `json:"transient,omitempty"`
}

type AudioTranscriptionsV1Output struct {
	Task     *string                         `json:"task,omitempty"`
	Language *string                         `json:"language,omitempty"`
	Duration *float64                        `json:"duration,omitempty"`
	Segments []AudioTranscriptionsV1Segments `json:"segments,omitempty"`
	Text     *string                         `json:"text,omitempty"`
	Error    *Error                          `json:"error,omitempty"`
}

func (impl *AudioTranscriptionsV1Output) GoString() string {
	buf := new(bytes.Buffer)
	json.NewEncoder(buf).Encode(impl)
	return buf.String()
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

	if input.ResponseFormat == nil {
		if err := writer.WriteField("response_format", "verbose_json"); err != nil {
			return nil, err
		}
	} else {
		switch *input.ResponseFormat {
		case "json", "verbose_json":
			if err := writer.WriteField("response_format", *input.ResponseFormat); err != nil {
				return nil, err
			}
		default:
			return nil, xerrors.Errorf("unsupport format: %s", *input.ResponseFormat)
		}
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
		return ret, nil
	case http.StatusUnauthorized:
		ret := &AudioTranscriptionsV1Output{}
		if err := json.NewDecoder(resp.Body).Decode(ret); err != nil {
			return nil, err
		}
		return ret, ErrUnauthorized
	case http.StatusBadGateway:
		buf := new(bytes.Buffer)
		io.Copy(buf, resp.Body)
		ret := &AudioTranscriptionsV1Output{
			Error: &Error{
				Message: buf.String(),
			},
		}
		return ret, xerrors.Errorf("msg: %s, error: %w", buf.String(), ErrStatusBadGateway)
	default:
		buf := new(bytes.Buffer)
		io.Copy(buf, resp.Body)
		ret := &AudioTranscriptionsV1Output{
			Error: &Error{
				Message: buf.String(),
			},
		}
		return ret, xerrors.Errorf("status_code: %d, msg: %s, error: %w", resp.StatusCode, buf.String(), ErrUnknown)
	}
}
