package openaiclaudeconverter

import (
	"testing"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/stretchr/testify/assert"
	"github.com/yanolja/ogem/openai"
	"github.com/yanolja/ogem/utils"
)
func TestToClaudeRequest(t *testing.T) {
	t.Run("valid request conversion", func(t *testing.T) {
		openaiRequest := &openai.ChatCompletionRequest{
			Model: "claude-3-sonnet",
			Messages: []openai.Message{
				{Role: "user", Content: &openai.MessageContent{String: utils.ToPtr("Hi there")}},
			},
			MaxTokens:   utils.ToPtr(int32(256)),
			Temperature: utils.ToPtr(float32(0.8)),
		}
		expected := &anthropic.MessageNewParams{
			Model:       anthropic.F("claude-3-sonnet-20240229"),
			MaxTokens:   anthropic.Int(256),
			Temperature: anthropic.F(float64(float32(0.8))),
			Messages: anthropic.F([]anthropic.MessageParam{
				{
					Role: anthropic.F(anthropic.MessageParamRoleUser),
					Content: anthropic.F([]anthropic.MessageParamContentUnion{
						anthropic.NewTextBlock("Hi there"),
					}),
				},
			}),
		}
		result, err := ToClaudeRequest(openaiRequest)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, expected, result)
	})

	t.Run("default MaxTokens based on model", func(t *testing.T) {
		openaiRequest := &openai.ChatCompletionRequest{
			Model: "claude-3-5-sonnet-20240620",
			Messages: []openai.Message{
				{Role: "user", Content: &openai.MessageContent{String: utils.ToPtr("Test message")}},
			},
		}

		result, err := ToClaudeRequest(openaiRequest)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, int64(8192), result.MaxTokens.Value)
	})

	t.Run("handles stop sequences", func(t *testing.T) {
		openaiRequest := &openai.ChatCompletionRequest{
			Model: "claude-3-sonnet",
			Messages: []openai.Message{
				{Role: "user", Content: &openai.MessageContent{String: utils.ToPtr("Hello")}},
			},
			StopSequences: &openai.StopSequences{Sequences: []string{"STOP"}},
		}

		result, err := ToClaudeRequest(openaiRequest)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, []string{"STOP"}, result.StopSequences.Value)
	})

	t.Run("handles system message", func(t *testing.T) {
		openaiRequest := &openai.ChatCompletionRequest{
			Model: "claude-3-sonnet",
			Messages: []openai.Message{
				{Role: "system", Content: &openai.MessageContent{String: utils.ToPtr("You are a helpful AI")}},
				{Role: "user", Content: &openai.MessageContent{String: utils.ToPtr("Hello")}},
			},
		}

		result, err := ToClaudeRequest(openaiRequest)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "You are a helpful AI", result.System.Value[0].Text.Value)
	})

	t.Run("handles TopP parameter", func(t *testing.T) {
		openaiRequest := &openai.ChatCompletionRequest{
			Model:  "claude-3-sonnet",
			TopP:   utils.ToPtr(float32(0.9)),
			Messages: []openai.Message{
				{Role: "user", Content: &openai.MessageContent{String: utils.ToPtr("How are you?")}},
			},
		}

		result, err := ToClaudeRequest(openaiRequest)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		tolerance := 1e-6
		assert.InDelta(t, 0.9, result.TopP.Value, tolerance)
	})

	t.Run("handles Tools and ToolChoice", func(t *testing.T) {
		tools := []openai.Tool{{Type: "function", Function: openai.FunctionTool{Name: "sum"}}}
		toolChoice := openai.ToolChoice{Value: utils.ToPtr(openai.ToolChoiceAuto)}

		openaiRequest := &openai.ChatCompletionRequest{
			Model:       "claude-3-sonnet",
			Tools:       tools,
			ToolChoice:  &toolChoice,
			Messages: []openai.Message{
				{Role: "user", Content: &openai.MessageContent{String: utils.ToPtr("Calculate sum")}},
			},
		}

		result, err := ToClaudeRequest(openaiRequest)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.Tools)
		assert.NotNil(t, result.ToolChoice)
	})

	t.Run("fails with missing messages", func(t *testing.T) {
		openaiRequest := &openai.ChatCompletionRequest{
			Model: "claude-3-sonnet",
		}

		result, err := ToClaudeRequest(openaiRequest)
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("empty request should fail", func(t *testing.T) {
		openaiRequest := &openai.ChatCompletionRequest{}

		result, err := ToClaudeRequest(openaiRequest)
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("unsupported response format", func(t *testing.T) {
		openaiRequest := &openai.ChatCompletionRequest{
			Model: "claude-3-sonnet",
			Messages: []openai.Message{
				{Role: "user", Content: &openai.MessageContent{String: utils.ToPtr("Hi there")}},
			},
			ResponseFormat: &openai.ResponseFormat{Type: "json"},
		}

		result, err := ToClaudeRequest(openaiRequest)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "response_format is not supported")
	})
}