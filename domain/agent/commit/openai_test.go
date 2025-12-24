package commit

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/responses"
	"github.com/openai/openai-go/v3/shared"
	"github.com/openai/openai-go/v3/shared/constant"
	"github.com/stretchr/testify/require"
	"github.com/yaroslav-koval/hange/domain/agent/entity"
)

type roundTripperFunc func(req *http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestOpenAICommitProcessor_GenCommitMessage(t *testing.T) {
	t.Parallel()

	commitData := entity.CommitData{
		UserInput:    "task context",
		Status:       "git status output",
		StagedStatus: "staged files",
		Diff:         "diff content",
	}

	t.Run("sends input and returns model output", func(t *testing.T) {
		t.Parallel()

		var capturedBody []byte

		rt := roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			body, err := io.ReadAll(req.Body)
			require.NoError(t, err)

			capturedBody = body
			require.Equal(t, "/responses", req.URL.Path)

			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{"Content-Type": []string{"application/json"}},
				Body:       io.NopCloser(bytes.NewReader(newStubResponse(t, "commit message"))),
			}, nil
		})

		client := openai.NewClient(
			option.WithBaseURL("http://example.com"),
			option.WithHTTPClient(&http.Client{Transport: rt}),
		)

		cp := &openAICommitProcessor{client: &client}

		msg, err := cp.GenCommitMessage(context.Background(), commitData)
		require.NoError(t, err)
		require.Equal(t, "commit message", msg)
		require.NotEmpty(t, capturedBody)

		expectedInput := cp.buildInput(commitData)

		var payload map[string]any
		require.NoError(t, json.Unmarshal(capturedBody, &payload))

		require.Equal(t, systemInstruction, payload["instructions"])
		require.Equal(t, string(commitModel), payload["model"])

		include, ok := payload["include"].([]any)
		require.True(t, ok)
		require.Contains(t, include, string(responses.ResponseIncludableFileSearchCallResults))

		input, ok := payload["input"].(string)
		require.True(t, ok)
		require.Equal(t, expectedInput, input)
	})

	t.Run("propagates request errors", func(t *testing.T) {
		t.Parallel()

		rt := roundTripperFunc(func(_ *http.Request) (*http.Response, error) {
			return nil, errors.New("request failed")
		})

		client := openai.NewClient(
			option.WithBaseURL("http://example.com"),
			option.WithHTTPClient(&http.Client{Transport: rt}),
		)

		cp := &openAICommitProcessor{client: &client}

		msg, err := cp.GenCommitMessage(context.Background(), commitData)
		require.Error(t, err)
		require.Empty(t, msg)
	})
}

func newStubResponse(t *testing.T, output string) []byte {
	t.Helper()

	resp := responses.Response{
		ID:                 "resp_123",
		CreatedAt:          1,
		Error:              responses.ResponseError{},
		IncompleteDetails:  responses.ResponseIncompleteDetails{},
		Instructions:       responses.ResponseInstructionsUnion{},
		Metadata:           shared.Metadata{},
		Model:              shared.ResponsesModel(commitModel),
		Object:             constant.Response("response"),
		Output:             []responses.ResponseOutputItemUnion{newOutputMessage(output)},
		ParallelToolCalls:  false,
		Temperature:        0,
		ToolChoice:         responses.ResponseToolChoiceUnion{},
		Tools:              []responses.ToolUnion{},
		TopP:               0,
		MaxOutputTokens:    80,
		Background:         false,
		Conversation:       responses.ResponseConversation{},
		MaxToolCalls:       0,
		PreviousResponseID: "",
		Prompt:             responses.ResponsePrompt{},
		PromptCacheKey:     "",
		PromptCacheRetention: responses.ResponsePromptCacheRetention(
			"",
		),
		Reasoning:        shared.Reasoning{},
		SafetyIdentifier: "",
		ServiceTier:      responses.ResponseServiceTier(""),
		Status:           responses.ResponseStatus("completed"),
		Text:             responses.ResponseTextConfig{},
		TopLogprobs:      0,
		Truncation:       responses.ResponseTruncation(""),
		Usage:            responses.ResponseUsage{},
		User:             "",
	}

	body, err := json.Marshal(resp)
	require.NoError(t, err)

	return body
}

func newOutputMessage(output string) responses.ResponseOutputItemUnion {
	return responses.ResponseOutputItemUnion{
		ID: "msg_1",
		Content: []responses.ResponseOutputMessageContentUnion{
			{
				Text: output,
				Type: string(constant.OutputText("output_text")),
			},
		},
		Role:   constant.Assistant("assistant"),
		Status: "completed",
		Type:   "message",
	}
}
