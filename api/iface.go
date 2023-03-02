package api

type OpenAIAPIIface interface {
	ListModelsV1() (*ListModelsV1Output, error)
}
