package api

type OpenAIAPIIface interface {
	ListModelsV1(*ListModelsV1Input) (*ListModelsV1Output, error)
}
