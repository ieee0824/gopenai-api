package api

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"reflect"

	"github.com/danielgtaylor/huma/schema"
	"github.com/samber/lo"
	"golang.org/x/xerrors"
)

type Message struct {
	Role    string `json:"role,omitempty"`
	Content string `json:"content,omitempty"`
}

type Tool struct {
	Type     string    `json:"type,omitempty"`
	Function *Function `json:"function,omitempty"`
}

// doc: https://platform.openai.com/docs/api-reference/chat
type ChatCompletionsV1Input struct {
	Model            *string     `json:"model,omitempty"`
	Messages         []Message   `json:"messages,omitempty"`
	Functions        []*Function `json:"functions,omitempty"` // Deprecated
	Temperature      *float32    `json:"temperature,omitempty"`
	TopP             *float32    `json:"top_p,omitempty"`
	N                int         `json:"n,omitempty"`
	Stop             []string    `json:"stop,omitempty"`
	MaxTokens        *int        `json:"max_tokens,omitempty"`
	PresencePenalty  *float32    `json:"presence_penalty,omitempty"`
	FrequencyPenalty *float32    `json:"frequency_penalty,omitempty"`
	LogitBias        any         `json:"logit_bias,omitempty"`
	User             *string     `json:"user,omitempty"`
	FunctionCall     any         `json:"function_call,omitempty"` // Deprecated
	Tools            []*Tool     `json:"tools,omitempty"`
	ToolChoice       any         `json:"tool_choice,omitempty"`
}

func (input *ChatCompletionsV1Input) Validate() error {
	if input.Model == nil {
		return xerrors.New("model is empty")
	}
	if len(input.Messages) == 0 {
		return xerrors.New("messages is empty")
	}
	return nil
}

type ChatCompletionsV1OutputUsage struct {
	PromptTokens     int `json:"prompt_tokens,omitempty"`
	CompletionTokens int `json:"completion_tokens,omitempty"`
	TotalTokens      int `json:"total_tokens,omitempty"`
}

type ChatCompletionsV1OutputChoiceFunctionCall struct {
	Name      string `json:"name,omitempty"`
	Arguments string `json:"arguments,omitempty"`
}

type ChatCompletionsV1OutputChoiceMessage struct {
	Role         string                                     `json:"role,omitempty"`
	Content      *string                                    `json:"content,omitempty"`
	FunctionCall *ChatCompletionsV1OutputChoiceFunctionCall `json:"function_call,omitempty"`
	ToolCalls    []ChatCompletionsV1OutputToolCall          `json:"tool_calls,omitempty"`
}

type ChatCompletionsV1OutputChoice struct {
	Message      ChatCompletionsV1OutputChoiceMessage `json:"message,omitempty"`
	FinishReason string                               `json:"finish_reason,omitempty"`
	Index        int                                  `json:"index,omitempty"`
}

type ChatCompletionsV1OutputToolCallFunction struct {
	Name      string `json:"name,omitempty"`
	Arguments string `json:"arguments,omitempty"`
}

type ChatCompletionsV1OutputToolCall struct {
	ID       string                                   `json:"id,omitempty"`
	Type     string                                   `json:"type,omitempty"`
	Function *ChatCompletionsV1OutputToolCallFunction `json:"function,omitempty"`
}

type ChatCompletionsV1Output struct {
	ID      *string                         `json:"id,omitempty"`
	Object  *string                         `json:"object,omitempty"`
	Created *int                            `json:"created,omitempty"`
	Model   *string                         `json:"model,omitempty"`
	Usage   *ChatCompletionsV1OutputUsage   `json:"usage,omitempty"`
	Choices []ChatCompletionsV1OutputChoice `json:"choices,omitempty"`
	Error   *Error                          `json:"error,omitempty"`
}

func (impl *ChatCompletionsV1Output) parseArgumentsFromFunctionCalls(funcName string, functionCalls []*ChatCompletionsV1OutputChoiceFunctionCall, v any) error {
	for _, fc := range functionCalls {
		if fc.Name != funcName {
			continue
		}
		if err := json.Unmarshal([]byte(fc.Arguments), v); err != nil {
			return err
		}
		return nil
	}
	return xerrors.Errorf("function name: %s is not found: %w", funcName, ErrParseFunctionCallingArguments)
}

func (impl *ChatCompletionsV1Output) parseArgumentsFromToolCalls(funcName string, toolCalls []ChatCompletionsV1OutputToolCall, v any) error {
	for _, tc := range toolCalls {
		if tc.Function.Name != funcName {
			continue
		}
		if err := json.Unmarshal([]byte(tc.Function.Arguments), v); err != nil {
			return err
		}
		return nil
	}
	return xerrors.Errorf("function name: %s is not found: %w", funcName, ErrParseFunctionCallingArguments)
}

// parse function calling arguments
func (impl *ChatCompletionsV1Output) ParseArguments(funcName string, v any) error {
	if impl.Choices == nil {
		return xerrors.Errorf("choices is nil: %w", ErrParseFunctionCallingArguments)
	}
	functionCalls := lo.Filter(lo.Map(impl.Choices, func(c ChatCompletionsV1OutputChoice, _ int) *ChatCompletionsV1OutputChoiceFunctionCall {
		return c.Message.FunctionCall
	}), func(fc *ChatCompletionsV1OutputChoiceFunctionCall, _ int) bool {
		return fc != nil
	})
	if len(functionCalls) != 0 {
		return impl.parseArgumentsFromFunctionCalls(funcName, functionCalls, v)
	}

	toolCalls := lo.Filter(lo.Map(impl.Choices, func(c ChatCompletionsV1OutputChoice, _ int) []ChatCompletionsV1OutputToolCall {
		return c.Message.ToolCalls
	}), func(tc []ChatCompletionsV1OutputToolCall, _ int) bool {
		return len(tc) != 0
	})

	for _, tc := range toolCalls {
		if err := impl.parseArgumentsFromToolCalls(funcName, tc, v); err == nil {
			return nil
		}
	}
	return xerrors.Errorf("function name: %s is not found: %w", funcName, ErrParseFunctionCallingArguments)
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
	case http.StatusBadGateway:
		buf := new(bytes.Buffer)
		io.Copy(buf, resp.Body)
		ret := &ChatCompletionsV1Output{
			Error: &Error{
				Message: buf.String(),
			},
		}
		return ret, xerrors.Errorf("msg: %s, error: %w", buf.String(), ErrStatusBadGateway)
	default:
		buf := new(bytes.Buffer)
		io.Copy(buf, resp.Body)
		ret := &ChatCompletionsV1Output{
			Error: &Error{
				Message: buf.String(),
			},
		}

		return ret, xerrors.Errorf("status_code: %d, msg: %s, error: %w", resp.StatusCode, buf.String(), ErrUnknown)
	}
}

// generate function_calling function
// Input the functionName, description, and return value type of the function as arguments.
// example:
//
//	type funcResult struct {
//	    Foo string `json:"foo"`
//	    Bar int    `json:"bar"`
//	    Baz bool   `json:"baz"`
//	}
//
// NewFunction("funcName", "description", funcResult{})
func NewFunction(funcName, description string, v any) (*Function, error) {
	parameters, err := schema.Generate(reflect.TypeOf(v))
	if err != nil {
		return nil, xerrors.Errorf("failed to generate schema: %w", err)
	}
	return &Function{
		Name:        funcName,
		Description: description,
		Paramaters:  parameters,
	}, nil

}

type Function struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Paramaters  *schema.Schema `json:"parameters"`
}
