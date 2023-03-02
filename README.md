# gopenai-api

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