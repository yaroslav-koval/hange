package explain

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/responses"
	"github.com/yaroslav-koval/hange/pkg/agent"
	"github.com/yaroslav-koval/hange/pkg/entities"
	"golang.org/x/sync/errgroup"
)

const explanationModel = openai.ChatModelGPT5Nano

var ErrTooManyAttempts = errors.New("too many attempts")
var ErrFailedToProcessFiles = errors.New("failed to process files")

func NewOpenAIExplainProcessor(client *openai.Client) agent.ExplainProcessor {
	return &explainProcessor{
		client: client,
		mutex:  &sync.RWMutex{},
	}
}

const explainInstruction = `You are a senior software engineer and codebase explainer.

Your task:
- You receive one or more files (source code, configs, docs, etc.).
- You must explain these files from a **developer’s perspective** to another developer.

How to respond:
1. Start with a **short high-level overview** of what the files collectively do.
2. Describe the **project structure**:
   - Focus on **folders**: what they contain, key responsibilities, and how they connect.
   - Only drill into **individual files** when they stand alone (e.g., the sole file in a folder or files at the root).
   - Call out **key structures** (types, classes, interfaces, functions, handlers, etc.) when relevant to that folder or single file.
3. Highlight:
   - Important **design decisions** or patterns.
   - Any **notable edge cases, constraints, or assumptions**.
   - How a new developer might **extend or modify** this code safely.

Style:
- Write as if you’re doing a **code walkthrough for a teammate**.
- Be clear, concise, and technical.
- If something is ambiguous, say that it’s unclear and explain **why** instead of guessing.
- Do **not** invent non-existent files, functions, or behavior.
- Do not suggest next activities.

You will be given the file names and their contents (possibly truncated). Base your explanation **only on the provided information**.`

const explainPrompt = `You are given the following project files. Explain them from a developer’s perspective.

Files: `

// TODO take values of files/vectorStore expiration from env

type explainProcessor struct {
	client      *openai.Client
	files       []*openai.FileObject
	vectorStore *openai.VectorStore
	mutex       *sync.RWMutex
}

func (ep *explainProcessor) ProcessFiles(ctx context.Context, files <-chan entities.File) error {
	if err := ep.uploadFiles(ctx, files); err != nil {
		return err
	}

	if err := ep.createVectorStore(ctx); err != nil {
		return err
	}

	return nil
}

func (ep *explainProcessor) uploadFiles(ctx context.Context, files <-chan entities.File) error {
	eg, ctx := errgroup.WithContext(ctx)

	consumed := false

	for !consumed {
		select {
		case <-ctx.Done():
			return fmt.Errorf("failed to upload files: %w", context.Canceled)
		default:
			f, ok := <-files
			if !ok {
				consumed = true
				break
			}

			eg.Go(func() error {
				params := openai.FileNewParams{
					File:    openai.File(bytes.NewReader(f.Data), f.Path, "text/plain"),
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
	}

	return eg.Wait()
}

func (ep *explainProcessor) createVectorStore(ctx context.Context) error {
	vs, err := ep.client.VectorStores.New(ctx, openai.VectorStoreNewParams{
		Name:             openai.String("hange_" + strconv.Itoa(int(time.Now().UTC().Unix()))),
		Metadata:         nil,
		ChunkingStrategy: openai.FileChunkingStrategyParamUnion{},
		ExpiresAfter: openai.VectorStoreNewParamsExpiresAfter{
			Days: 1,
		},
	})
	if err != nil {
		return err
	}

	slog.Info("Waiting for vector store processing...")

	vs, err = retry(
		ctx,
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

	ep.mutex.Lock()
	ep.vectorStore = vs
	ep.mutex.Unlock()

	slog.Info("Vector store created")
	slog.Debug(vs.RawJSON())

	ep.mutex.RLock()

	fileIDs := make([]string, len(ep.files))
	for i, f := range ep.files {
		fileIDs[i] = f.ID
	}

	filesCount := int64(len(ep.files))

	ep.mutex.RUnlock()

	_, err = ep.client.VectorStores.FileBatches.New(ctx, vs.ID, openai.VectorStoreFileBatchNewParams{
		FileIDs: fileIDs,
	})
	if err != nil {
		return err
	}

	slog.Info("Started files batch processing...")

	// wait until files are processed
	vs, err = retry(ctx, func() (*openai.VectorStore, bool, error) {
		vs, err = ep.client.VectorStores.Get(ctx, ep.vectorStore.ID)
		if err != nil {
			return nil, false, err
		}

		slog.Debug(fmt.Sprintf("Files processing status: %v\n%s\n", vs.Status, vs.FileCounts.RawJSON()))

		if (vs.FileCounts.Total - vs.FileCounts.InProgress) == filesCount {
			return vs, true, nil
		}

		return nil, false, nil
	}, time.Second, 0)
	if err != nil {
		return err
	}

	if vs.FileCounts.Failed != 0 {
		return ErrFailedToProcessFiles
	}

	slog.Info("File batch is uploaded to vector store")

	return nil
}

// 0 attempts means infinite polling
func retry[T any](ctx context.Context, f func() (*T, bool, error), interval time.Duration, attempts int) (*T, error) {
	attemptsCounter := 0

	for {
		select {
		case <-ctx.Done():
			return nil, context.Canceled
		default:
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
}

func (ep *explainProcessor) vectorStoreID() string {
	ep.mutex.RLock()
	defer ep.mutex.RUnlock()

	return ep.vectorStore.ID
}

func (ep *explainProcessor) Cleanup(ctx context.Context) {
	wg := &sync.WaitGroup{}

	// should do cleanup even if context in cancelled
	ctx = context.Background()

	ep.mutex.RLock()
	defer ep.mutex.RUnlock()

	for _, f := range ep.files {
		wg.Go(func() {
			ep.mutex.RLock()
			defer ep.mutex.RUnlock()

			_, err := ep.client.Files.Delete(ctx, f.ID)
			if err != nil {
				slog.Error(fmt.Sprintf("Failed to delete file by id %s: %s", f.ID, err))
			} else {
				slog.Debug(fmt.Sprintf("File is deleted by id %s", f.ID))
			}
		})
	}

	wg.Go(func() {
		ep.mutex.RLock()
		defer ep.mutex.RUnlock()

		if ep.vectorStore == nil {
			return
		}

		_, err := ep.client.VectorStores.Delete(ctx, ep.vectorStore.ID)
		if err != nil {
			slog.Error(fmt.Sprintf("Failed to delete vector store by id %s: %s", ep.vectorStore.ID, err))
		} else {
			slog.Debug(fmt.Sprintf("Vector store is deleted by id %s", ep.vectorStore.ID))
		}
	})

	wg.Wait()

	slog.Info("Data cleanup is finished")
}

func (ep *explainProcessor) ExecuteExplainRequest(ctx context.Context) (string, error) {
	ep.mutex.RLock()

	fileNames := make([]string, len(ep.files))
	for i, f := range ep.files {
		fileNames[i] = f.Filename
	}

	ep.mutex.RUnlock()

	input := explainPrompt + strings.Join(fileNames, ", ")

	slog.Info("Calling explanation model...")

	resp, err := ep.client.Responses.New(ctx, responses.ResponseNewParams{
		Instructions: openai.String(explainInstruction),
		Include: []responses.ResponseIncludable{
			responses.ResponseIncludableFileSearchCallResults,
		},
		Input: responses.ResponseNewParamsInputUnion{OfString: openai.String(input)},
		Model: explanationModel,
		Tools: []responses.ToolUnionParam{
			{
				OfFileSearch: &responses.FileSearchToolParam{
					VectorStoreIDs: []string{ep.vectorStoreID()},
				},
			},
		},
	})
	if err != nil {
		return "", err
	}

	return resp.OutputText(), nil
}
