package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"html/template"

	openai "github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
)

var scorePrompt = `You are a URL classifier for online stores. Your job is to rate a URL on how likely it is to lead to a fake online store. You can use values ​​between 1 and 10, where low values ​​mean low probability and high values ​​mean high probability.

Domain:
{{ .Domain }}
`

func GetDomainScore(client openai.Client, domain string) (int, error) {
	tmpl, err := template.New("scorePrompt").Parse(scorePrompt)
	if err != nil {
		return 0, err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, map[string]string{"Domain": domain}); err != nil {
		return 0, err
	}

	// Structured output JSON schema definition
	type Response struct {
		Score int `json:"score"`
	}
	var response Response
	schema, err := jsonschema.GenerateSchemaForType(response)
	if err != nil {
		return 0, err
	}

	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:       openai.GPT4oMini,
			MaxTokens:   300,
			Temperature: 0.0,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: buf.String(),
				},
			},
			ResponseFormat: &openai.ChatCompletionResponseFormat{
				Type: openai.ChatCompletionResponseFormatTypeJSONSchema,
				JSONSchema: &openai.ChatCompletionResponseFormatJSONSchema{
					Name:   "score",
					Schema: schema,
					Strict: true,
				},
			},
		},
	)
	if err != nil {
		return 0, err
	}
	content := resp.Choices[0].Message.Content
	if err := json.Unmarshal([]byte(content), &response); err != nil {
		return 0, err
	}
	return response.Score, nil
}
