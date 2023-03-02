package api

type OpenAIAPIIface interface {
	ListModelsV1(*ListModelsV1Input) (*ListModelsV1Output, error)
	ChatCompletionsV1(input *ChatCompletionsV1Input) (*ChatCompletionsV1Output, error)
}
