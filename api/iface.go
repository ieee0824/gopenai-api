package api

type OpenAIAPIIface interface {
	ListModelsV1(*ListModelsV1Input) (*ListModelsV1Output, error)
	ChatCompletionsV1(input *ChatCompletionsV1Input) (*ChatCompletionsV1Output, error)
	AudioTranscriptionsV1(input *AudioTranscriptionsV1Input) (*AudioTranscriptionsV1Output, error)
	ListFileV1(input *ListFileV1Input) (*ListFileV1Output, error)
	ImagesGenerationsV1(input *ImagesGenerationsV1Input) (*ImagesGenerationsV1Output, error)
}
