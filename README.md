# gopenai-api

### install
```shell
go get github.com/ieee0824/gopenai-api
```

### sample
```Go
package main

import (
	"fmt"

	"github.com/ieee0824/gopenai-api/api"
	"github.com/ieee0824/gopenai-api/config"
	"github.com/samber/lo"
)

func main() {
	a := api.New(&config.Configuration{
		ApiKey:       lo.ToPtr("api-key"),
		Organization: lo.ToPtr("organization-id"),
	})

	fmt.Println(a.ChatCompletionsV1(&api.ChatCompletionsV1Input{
		Model: lo.ToPtr("gpt-3.5-turbo"),
		Messages: []api.Message{
			{
				Role:    "user",
				Content: "ChatGPT 3.5のapiの使い方を教えてください",
			},
		},
	}))
}
```

### function calling sample
```Go
package main

import (
	"fmt"

	"github.com/ieee0824/gopenai-api/api"
	"github.com/ieee0824/gopenai-api/config"
)

type weatherAPIReq struct {
	Location string `json:"location"`
}

func main() {
	model := "gpt-3.5-turbo"
	apiKey := "api-key"
	org := "org-id"
	const funcName = "weather"

	mf, err := api.NewFunction(
		funcName,
		"",
		weatherAPIReq{},
	)
	if err != nil {
		panic(err)
	}

	ai := api.New(&config.Configuration{
		ApiKey:       &apiKey,
		Organization: &org,
	})

	result, err := ai.ChatCompletionsV1(&api.ChatCompletionsV1Input{
		Model: &model,
		Messages: []api.Message{
			{
				Role:    "user",
				Content: "間ノ岳の天気を教えてください",
			},
		},
		Functions: []*api.Function{
			mf,
		},
	})

	if err != nil {
		panic(err)
	}

	ret := &weatherAPIReq{}
	if err := result.ParseArguments(funcName, ret); err != nil {
		panic(err)
	}

	fmt.Println(ret)
}
```