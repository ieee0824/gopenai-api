package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

type Message struct {
	Role    string `json:"role,omitempty"`
	Content string `json:"content,omitempty"`
}

// doc: https://platform.openai.com/docs/api-reference/chat
type ChatCompletionsV1Input struct {
	Model            *string   `json:"model,omitempty"`
	Messages         []Message `json:"messages,omitempty"`
	Temperature      *float32  `json:"temperature,omitempty"`
	TopP             *float32  `json:"top_p,omitempty"`
	N                int       `json:"n,omitempty"`
	Stop             []string  `json:"stop,omitempty"`
	MaxTokens        *int      `json:"max_tokens,omitempty"`
	PresencePenalty  *float32  `json:"presence_penalty,omitempty"`
	FrequencyPenalty *float32  `json:"frequency_penalty,omitempty"`
	LogitBias        any       `json:"logit_bias,omitempty"`
	User             *string   `json:"user,omitempty"`
}

func (input *ChatCompletionsV1Input) Validate() error {
	if input.Model == nil {
		return errors.New("model is empty")
	}
	if len(input.Messages) == 0 {
		return errors.New("messages is empty")
	}
	return nil
}

type ChatCompletionsV1OutputUsage struct {
	PromptTokens     int `json:"prompt_tokens,omitempty"`
	CompletionTokens int `json:"completion_tokens,omitempty"`
	TotalTokens      int `json:"total_tokens,omitempty"`
}

type ChatCompletionsV1OutputChoiceMessage struct {
	Role    string `json:"role,omitempty"`
	Content string `json:"content,omitempty"`
}

type ChatCompletionsV1OutputChoice struct {
	Message      ChatCompletionsV1OutputChoiceMessage `json:"message,omitempty"`
	FinishReason string                               `json:"finish_reason,omitempty"`
	Index        int                                  `json:"index,omitempty"`
}

type ChatCompletionsV1Output struct {
	ID      *string                         `json:"id,omitempty"`
	Object  *string                         `json:"object,omitempty"`
	Created *int                            `json:"created,omitempty"`
	Model   *string                         `json:"model,omitempty"`
	Usage   *ChatCompletionsV1OutputUsage   `json:"usage,omitempty"`
	Choices []ChatCompletionsV1OutputChoice `json:"choices,omitempty"`

	Error *Error `json:"error,omitempty"`
}

func (impl *ChatCompletionsV1Output) String() string {
	buf := new(bytes.Buffer)
	json.NewEncoder(buf).Encode(impl)
	return buf.String()
}

func (api *OpenAIAPI) ChatCompletionsV1(input *ChatCompletionsV1Input) (*ChatCompletionsV1Output, error) {
	if err := input.Validate(); err != nil {
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
	endpoint.Path = "/v1/chat/completions"
	req, err := http.NewRequest(
		http.MethodPost,
		endpoint.String(),
		reqBody,
	)
	req.Header.Add("Content-Type", "application/json")
	api.setToken(req)
	if err != nil {
		return nil, err
	}
	resp, err := api.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		ret := &ChatCompletionsV1Output{}
		if err := json.NewDecoder(resp.Body).Decode(ret); err != nil {
			return nil, err
		}
		return ret, nil
	case http.StatusUnauthorized:
		ret := &ChatCompletionsV1Output{}
		if err := json.NewDecoder(resp.Body).Decode(ret); err != nil {
			return nil, err
		}
		return ret, ErrUnauthorized
	default:
		buf := new(bytes.Buffer)
		io.Copy(buf, resp.Body)
		ret := &ChatCompletionsV1Output{
			Error: &Error{
				Message: buf.String(),
			},
		}
		return ret, ErrUnknown
	}

	return nil, nil
}
