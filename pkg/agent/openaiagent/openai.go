package openaiagent

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/responses"
	"github.com/yaroslav-koval/hange/pkg/agent"
	"github.com/yaroslav-koval/hange/pkg/auth"
	"golang.org/x/sync/errgroup"
)

var ErrTooManyAttempts = errors.New("too many attempts")

func NewOpenAIUseCase(auth auth.Auth) agent.AIUseCase {
	return &openAIUseCase{
		auth: auth,
	}
}

type openAIUseCase struct {
	auth auth.Auth
}

type File struct {
	Name      string
	Extension string
	Data      io.Reader
}

func (f *File) FullName() string {
	return fmt.Sprintf("%s.%s", f.Name, f.Extension)
}

func (o *openAIUseCase) ExplainFiles(ctx context.Context, files []File) (string, error) {
	token, err := o.auth.GetToken()
	if err != nil {
		return "", err
	}

	client := openai.NewClient(
		option.WithAPIKey(token),
	)

	ep := newExplainProcessor(&client)
	defer ep.CleanupData(ctx) // TODO uncomment

	if err = ep.UploadFiles(ctx, files); err != nil {
		return "", err
	}

	fileNames := make([]string, len(files))
	for i, f := range files {
		fileNames[i] = f.FullName()
	}

	input := explainPrompt + strings.Join(fileNames, ", ")

	resp, err := client.Responses.New(ctx, responses.ResponseNewParams{
		//Background:           param.Opt[bool]{},
		Instructions: openai.String(explainInstruction),
		//MaxOutputTokens:      param.Opt[int64]{},
		//MaxToolCalls:         param.Opt[int64]{},
		//ParallelToolCalls:    param.Opt[bool]{},
		//PreviousResponseID:   param.Opt[string]{},
		//Store:                param.Opt[bool]{},
		//Temperature:          param.Opt[float64]{},
		//TopLogprobs:          param.Opt[int64]{},
		//TopP:                 param.Opt[float64]{},
		//PromptCacheKey:       param.Opt[string]{},
		//SafetyIdentifier:     param.Opt[string]{},
		//User:                 param.Opt[string]{},
		//Conversation:         responses.ResponseNewParamsConversationUnion{},
		Include: []responses.ResponseIncludable{
			responses.ResponseIncludableFileSearchCallResults,
		},
		//Metadata:             nil,
		//Prompt:               responses.ResponsePromptParam{},
		//PromptCacheRetention: "",
		//ServiceTier:          "",
		//StreamOptions:        responses.ResponseNewParamsStreamOptions{},
		//Truncation: "",
		Input: responses.ResponseNewParamsInputUnion{
			OfString: openai.String(input),
		},
		Model: "gpt-5.1", // "gpt-5.1-codex-max"
		//Reasoning:  shared.ReasoningParam{},
		//Text:       responses.ResponseTextConfigParam{},
		//ToolChoice: responses.ResponseNewParamsToolChoiceUnion{},
		Tools: []responses.ToolUnionParam{
			{
				OfFileSearch: &responses.FileSearchToolParam{
					VectorStoreIDs: []string{ep.VectorStoreID()},
				},
			},
			//{ TODO research
			//	OfCodeInterpreter:
			//},
		},
	})
	if err != nil {
		return "", err
	}

	slog.Info(fmt.Sprintf("Raw response from LLM:%s\n", resp.RawJSON()))

	return "", nil
}

const explainInstruction = `You are a senior software engineer and codebase explainer.

Your task:
- You receive one or more files (source code, configs, docs, etc.).
- You must explain these files from a **developer’s perspective** to another developer.

How to respond:
1. Start with a **short high-level overview** of what the files collectively do.
2. Then go **file by file**:
   - Explain the **purpose** of each file.
   - Describe **key structures** (types, classes, interfaces, functions, handlers, etc.).
   - Explain **how the file fits into the overall system** (dependencies, entry points, responsibilities).
3. Highlight:
   - Important **design decisions** or patterns.
   - Any **notable edge cases, constraints, or assumptions**.
   - How a new developer might **extend or modify** this code safely.

Style:
- Write as if you’re doing a **code walkthrough for a teammate**.
- Be clear, concise, and technical.
- If something is ambiguous, say that it’s unclear and explain **why** instead of guessing.
- Do **not** invent non-existent files, functions, or behavior.

You will be given the file names and their contents (possibly truncated). Base your explanation **only on the provided information**.`

// TODO add project context
const explainPrompt = `You are given the following project files. Explain them from a developer’s perspective.

Files: `

// TODO take values of expiration from env

type explainProcessor struct {
	client      *openai.Client
	files       []*openai.FileObject
	vectorStore *openai.VectorStore
	mutex       *sync.RWMutex
}

func newExplainProcessor(client *openai.Client) *explainProcessor {
	return &explainProcessor{
		client: client,
		mutex:  &sync.RWMutex{},
	}
}

func (ep *explainProcessor) UploadFiles(ctx context.Context, files []File) error {
	eg := &errgroup.Group{}

	for _, f := range files {
		eg.Go(func() error {
			params := openai.FileNewParams{
				File:    openai.File(f.Data, f.FullName(), "text/plain"),
				Purpose: openai.FilePurposeUserData,
			}

			// TODO make a fix PR in SDK. ExpiresAfter bug in SDK. It aligns field by dot: "expires_after.anchor: created_at"
			// https://github.com/openai/openai-go/issues/563?utm_source=chatgpt.com
			// params.ExpiresAfter = openai.FileNewParamsExpiresAfter{
			//   Seconds: 1 * hourInSeconds,
			// }

			// expiration is needed to avoid user's manual cleanup
			params.SetExtraFields(map[string]any{
				"expires_after[anchor]":  "created_at",
				"expires_after[seconds]": strconv.Itoa(60 * 60),
			})

			fileResp, err := ep.client.Files.New(ctx, params)
			if err != nil {
				return err
			}

			slog.Debug(fmt.Sprintf("File created:\n%s\n", fileResp.RawJSON()))

			ep.mutex.Lock()
			defer ep.mutex.Unlock()
			ep.files = append(ep.files, fileResp)

			return nil
		})
	}

	err := eg.Wait()
	if err != nil {
		return err
	}

	if err = ep.createVectorStore(ctx); err != nil {
		return err
	}

	return nil
}

func (ep *explainProcessor) createVectorStore(ctx context.Context) error {
	ep.mutex.RLock()

	fileIDs := make([]string, len(ep.files))
	for i, f := range ep.files {
		fileIDs[i] = f.ID
	}

	ep.mutex.RUnlock()

	vs, err := ep.client.VectorStores.New(ctx, openai.VectorStoreNewParams{
		Name:             openai.String("hange_" + strconv.Itoa(int(time.Now().UTC().Unix()))),
		Metadata:         nil,
		ChunkingStrategy: openai.FileChunkingStrategyParamUnion{},
		ExpiresAfter: openai.VectorStoreNewParamsExpiresAfter{
			Days: 1,
		},
		//FileIDs: fileIDs,
	})
	if err != nil {
		return err
	}

	slog.Info("Waiting for vector store processing...")

	vs, err = retry(
		func() (*openai.VectorStore, bool, error) {
			vecStore, err := ep.client.VectorStores.Get(ctx, vs.ID)
			if err != nil {
				return nil, false, err
			}

			if vecStore.Status == openai.VectorStoreStatusInProgress {
				return nil, false, nil
			}

			return vecStore, true, nil
		},
		500*time.Millisecond,
		5)
	if err != nil {
		return err
	}

	slog.Debug(fmt.Sprintf("Vector store created:\n%s\n", vs.RawJSON()))

	_, err = ep.client.VectorStores.FileBatches.NewAndPoll(ctx, vs.ID, openai.VectorStoreFileBatchNewParams{
		FileIDs: fileIDs,
	}, 0)

	slog.Debug(fmt.Sprintf("File batch is uploaded to vector store %s:\n", vs.ID))

	ep.mutex.Lock()
	defer ep.mutex.Unlock()
	ep.vectorStore = vs

	return nil
}

func retry[T any](f func() (*T, bool, error), interval time.Duration, attempts int) (*T, error) {
	attemptsCounter := 0

	for {
		v, ok, err := f()
		if err != nil {
			return nil, err
		}

		if ok {
			return v, nil
		}

		attemptsCounter++
		if attemptsCounter == attempts {
			return nil, ErrTooManyAttempts
		}

		time.Sleep(interval)
	}
}

func (ep *explainProcessor) VectorStoreID() string {
	ep.mutex.RLock()
	defer ep.mutex.RUnlock()

	return ep.vectorStore.ID
}

func (ep *explainProcessor) CleanupData(ctx context.Context) {
	wg := &sync.WaitGroup{}

	for _, f := range ep.files {
		wg.Go(func() {
			ep.mutex.RLock()
			defer ep.mutex.RUnlock()

			_, err := ep.client.Files.Delete(ctx, f.ID)
			if err != nil {
				slog.Debug(fmt.Sprintf("Failed to delete file by id %s: %s", f.ID, err))
			} else {
				slog.Debug(fmt.Sprintf("File is deleted by id %s", f.ID))
			}
		})
	}

	wg.Go(func() {
		ep.mutex.RLock()
		defer ep.mutex.RUnlock()

		_, err := ep.client.VectorStores.Delete(ctx, ep.vectorStore.ID)
		if err != nil {
			slog.Debug(fmt.Sprintf("Failed to delete vector store by id %s: %s", ep.vectorStore.ID, err))
		} else {
			slog.Debug(fmt.Sprintf("Vector store is deleted by id %s", ep.vectorStore.ID))
		}
	})

	wg.Wait()

	slog.Debug("Cleanup is finished")
}
