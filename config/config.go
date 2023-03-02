package config

type Configuration struct {
	ApiKey       *string
	Organization *string
	Endpoint     *string //default: api.openai.com
}
