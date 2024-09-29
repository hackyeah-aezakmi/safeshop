package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"html/template"
	"log"

	openai "github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
)

var scorePrompt = `You are an expert URL classifier specializing in detecting fraudulent online stores. Your task is to analyze the given domain and assess the likelihood that it leads to a fake online store. 

Consider the following factors in your analysis:
1. Domain name structure and relevance to legitimate businesses
2. Use of suspicious keywords or misspellings
3. Presence of well-known brand names in unexpected contexts
4. Unusual top-level domains (TLDs) that are often associated with scams

Rate the domain on a scale from 1 to 5:
1: Very low probability of being a fake store
2: Low probability
3: Moderate probability
4: High probability
5: Very high probability of being a fake store

Domain:
{{ .Domain }}
`

func GetDomainScore(client openai.Client, domain string) (float64, error) {
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

	log.Printf("Domain %s score by AI: %d", domain, response.Score)
	// Convert score from 1-5 to 0-1
	score := (float64(response.Score) - 1) / 4.0
	return score, nil
}
