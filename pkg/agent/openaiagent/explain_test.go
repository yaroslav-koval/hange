package openaiagent

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/responses"
	"github.com/openai/openai-go/v3/shared"
	"github.com/openai/openai-go/v3/shared/constant"
	"github.com/stretchr/testify/require"
	"github.com/yaroslav-koval/hange/pkg/agent"
)

func TestExplainProcessor_uploadFiles(t *testing.T) {
	t.Parallel()

	var (
		mu      sync.Mutex
		counter int
	)

	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "/files", r.URL.Path)

		mu.Lock()
		counter++
		id := fmt.Sprintf("file_%d", counter)
		mu.Unlock()

		resp := openai.FileObject{
			ID:        id,
			Bytes:     1,
			CreatedAt: time.Now().UTC().Unix(),
			Filename:  fmt.Sprintf("%s.txt", id),
			Object:    constant.File("file"),
			Purpose:   openai.FileObjectPurposeUserData,
			Status:    openai.FileObjectStatusProcessed,
		}

		w.Header().Set("Content-Type", "application/json")
		require.NoError(t, json.NewEncoder(w).Encode(resp))
	})

	filesCh := make(chan agent.File, 2)
	filesCh <- agent.File{Name: "one.go", Data: strings.NewReader("one")}
	filesCh <- agent.File{Name: "two.md", Data: strings.NewReader("two")}
	close(filesCh)

	ep := newTestExplainProcessor(client)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	err := ep.uploadFiles(ctx, filesCh)
	require.NoError(t, err)

	require.Len(t, ep.files, 2)

	var ids []string
	for _, f := range ep.files {
		ids = append(ids, f.ID)
	}
	require.ElementsMatch(t, []string{"file_1", "file_2"}, ids)

	mu.Lock()
	require.Equal(t, 2, counter)
	mu.Unlock()
}

func TestExplainProcessor_uploadFiles_contextCancelled(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	filesCh := make(chan agent.File)
	close(filesCh)

	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
	})

	ep := newTestExplainProcessor(client)

	err := ep.uploadFiles(ctx, filesCh)
	require.ErrorIs(t, err, ErrContextCancelled)
	require.Empty(t, ep.files)
}

func TestExplainProcessor_createVectorStore(t *testing.T) {
	t.Parallel()

	testFiles := []*openai.FileObject{
		{ID: "file_a"},
		{ID: "file_b"},
	}

	var receivedFileIDs []string

	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/vector_stores":
			w.Header().Set("Content-Type", "application/json")
			require.NoError(t, json.NewEncoder(w).Encode(newVectorStore("vs_test", openai.VectorStoreStatusInProgress, len(testFiles))))
		case r.Method == http.MethodGet && r.URL.Path == "/vector_stores/vs_test":
			w.Header().Set("Content-Type", "application/json")
			require.NoError(t, json.NewEncoder(w).Encode(newVectorStore("vs_test", openai.VectorStoreStatusCompleted, len(testFiles))))
		case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/file_batches"):
			var body struct {
				FileIDs []string `json:"file_ids"`
			}
			require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
			receivedFileIDs = append(receivedFileIDs, body.FileIDs...)

			resp := openai.VectorStoreFileBatch{
				ID:        "vs_batch_1",
				CreatedAt: time.Now().UTC().Unix(),
				FileCounts: openai.VectorStoreFileBatchFileCounts{
					Cancelled:  0,
					Completed:  int64(len(body.FileIDs)),
					Failed:     0,
					InProgress: 0,
					Total:      int64(len(body.FileIDs)),
				},
				Object:        constant.VectorStoreFilesBatch("vector_store.file_batch"),
				Status:        openai.VectorStoreFileBatchStatusCompleted,
				VectorStoreID: "vs_test",
			}

			w.Header().Set("Content-Type", "application/json")
			require.NoError(t, json.NewEncoder(w).Encode(resp))
		default:
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
	})

	ep := newTestExplainProcessor(client)
	ep.files = testFiles

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	err := ep.createVectorStore(ctx)
	require.NoError(t, err)
	require.NotNil(t, ep.vectorStore)
	require.Equal(t, "vs_test", ep.vectorStore.ID)
	require.ElementsMatch(t, []string{"file_a", "file_b"}, receivedFileIDs)
}

func TestExplainProcessor_createVectorStore_waitsForFileProcessing(t *testing.T) {
	t.Parallel()

	testFiles := []*openai.FileObject{
		{ID: "file_a"},
		{ID: "file_b"},
	}

	var (
		batchCreated    bool
		getRequests     int
		receivedFileIDs []string
	)

	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/vector_stores":
			w.Header().Set("Content-Type", "application/json")
			require.NoError(t, json.NewEncoder(w).Encode(newVectorStore("vs_retry", openai.VectorStoreStatusInProgress, len(testFiles))))
		case r.Method == http.MethodGet && r.URL.Path == "/vector_stores/vs_retry":
			getRequests++

			w.Header().Set("Content-Type", "application/json")
			resp := newVectorStore("vs_retry", openai.VectorStoreStatusCompleted, len(testFiles))

			switch {
			case !batchCreated && getRequests == 1:
				resp.Status = openai.VectorStoreStatusInProgress
			case batchCreated && getRequests == 3:
				resp.FileCounts.InProgress = int64(len(testFiles))
				resp.FileCounts.Completed = 0
			}

			require.NoError(t, json.NewEncoder(w).Encode(resp))
		case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/file_batches"):
			batchCreated = true

			var body struct {
				FileIDs []string `json:"file_ids"`
			}
			require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
			receivedFileIDs = append(receivedFileIDs, body.FileIDs...)

			resp := openai.VectorStoreFileBatch{
				ID:        fmt.Sprintf("batch_%d", len(receivedFileIDs)),
				CreatedAt: time.Now().UTC().Unix(),
				FileCounts: openai.VectorStoreFileBatchFileCounts{
					Cancelled:  0,
					Completed:  0,
					Failed:     0,
					InProgress: int64(len(body.FileIDs)),
					Total:      int64(len(body.FileIDs)),
				},
				Object:        constant.VectorStoreFilesBatch("vector_store.file_batch"),
				Status:        openai.VectorStoreFileBatchStatusInProgress,
				VectorStoreID: "vs_retry",
			}

			w.Header().Set("Content-Type", "application/json")
			require.NoError(t, json.NewEncoder(w).Encode(resp))
		default:
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
	})

	ep := newTestExplainProcessor(client)
	ep.files = testFiles

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := ep.createVectorStore(ctx)
	require.NoError(t, err)
	require.NotNil(t, ep.vectorStore)
	require.Equal(t, "vs_retry", ep.vectorStore.ID)
	require.ElementsMatch(t, []string{"file_a", "file_b"}, receivedFileIDs)
	require.GreaterOrEqual(t, getRequests, 4)
}

func TestExplainProcessor_createVectorStore_returnsErrorOnFailedFiles(t *testing.T) {
	t.Parallel()

	testFiles := []*openai.FileObject{
		{ID: "file_a"},
		{ID: "file_b"},
	}

	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/vector_stores":
			w.Header().Set("Content-Type", "application/json")
			require.NoError(t, json.NewEncoder(w).Encode(newVectorStore("vs_fail", openai.VectorStoreStatusCompleted, len(testFiles))))
		case r.Method == http.MethodGet && r.URL.Path == "/vector_stores/vs_fail":
			resp := newVectorStore("vs_fail", openai.VectorStoreStatusCompleted, len(testFiles))
			resp.FileCounts.Failed = 1
			resp.FileCounts.Completed = int64(len(testFiles) - 1)
			resp.FileCounts.InProgress = 0

			w.Header().Set("Content-Type", "application/json")
			require.NoError(t, json.NewEncoder(w).Encode(resp))
		case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/file_batches"):
			resp := openai.VectorStoreFileBatch{
				ID:        "vs_fail_batch",
				CreatedAt: time.Now().UTC().Unix(),
				FileCounts: openai.VectorStoreFileBatchFileCounts{
					Cancelled:  0,
					Completed:  0,
					Failed:     0,
					InProgress: int64(len(testFiles)),
					Total:      int64(len(testFiles)),
				},
				Object:        constant.VectorStoreFilesBatch("vector_store.file_batch"),
				Status:        openai.VectorStoreFileBatchStatusInProgress,
				VectorStoreID: "vs_fail",
			}

			w.Header().Set("Content-Type", "application/json")
			require.NoError(t, json.NewEncoder(w).Encode(resp))
		default:
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
	})

	ep := newTestExplainProcessor(client)
	ep.files = testFiles

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := ep.createVectorStore(ctx)
	require.ErrorIs(t, err, ErrFailedToProcessFiles)
}

func TestExplainProcessor_createVectorStore_returnsErrWhenVectorStoreStuck(t *testing.T) {
	t.Parallel()

	testFiles := []*openai.FileObject{
		{ID: "file_a"},
	}

	var getCalls int

	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/vector_stores":
			w.Header().Set("Content-Type", "application/json")
			require.NoError(t, json.NewEncoder(w).Encode(newVectorStore("vs_stuck", openai.VectorStoreStatusInProgress, len(testFiles))))
		case r.Method == http.MethodGet && r.URL.Path == "/vector_stores/vs_stuck":
			getCalls++

			w.Header().Set("Content-Type", "application/json")
			require.NoError(t, json.NewEncoder(w).Encode(newVectorStore("vs_stuck", openai.VectorStoreStatusInProgress, len(testFiles))))
		default:
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
	})

	ep := newTestExplainProcessor(client)
	ep.files = testFiles

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := ep.createVectorStore(ctx)
	require.ErrorIs(t, err, ErrTooManyAttempts)
	require.Nil(t, ep.vectorStore)
	require.Equal(t, 5, getCalls)
}

func TestExplainProcessor_Cleanup(t *testing.T) {
	t.Parallel()

	var (
		mu                 sync.Mutex
		deletedFiles       = map[string]int{}
		vectorStoreDeleted bool
		wg                 sync.WaitGroup
	)

	wg.Add(3)

	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodDelete && strings.HasPrefix(r.URL.Path, "/files/"):
			id := strings.TrimPrefix(r.URL.Path, "/files/")

			mu.Lock()
			deletedFiles[id]++
			mu.Unlock()
			wg.Done()

			resp := openai.FileDeleted{
				ID:      id,
				Deleted: true,
				Object:  constant.File("file"),
			}

			w.Header().Set("Content-Type", "application/json")
			require.NoError(t, json.NewEncoder(w).Encode(resp))
		case r.Method == http.MethodDelete && strings.HasPrefix(r.URL.Path, "/vector_stores/"):
			id := strings.TrimPrefix(r.URL.Path, "/vector_stores/")
			require.Equal(t, "vs_cleanup", id)

			mu.Lock()
			vectorStoreDeleted = true
			mu.Unlock()
			wg.Done()

			resp := openai.VectorStoreDeleted{
				ID:      id,
				Deleted: true,
				Object:  constant.VectorStoreDeleted("vector_store.deleted"),
			}

			w.Header().Set("Content-Type", "application/json")
			require.NoError(t, json.NewEncoder(w).Encode(resp))
		default:
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
	})

	ep := newTestExplainProcessor(client)
	ep.files = []*openai.FileObject{
		{ID: "file_a"},
		{ID: "file_b"},
	}
	ep.vectorStore = &openai.VectorStore{ID: "vs_cleanup"}

	ep.Cleanup(context.Background())

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatalf("cleanup did not finish requests in time")
	}

	mu.Lock()
	defer mu.Unlock()

	require.True(t, vectorStoreDeleted)
	require.Equal(t, map[string]int{"file_a": 1, "file_b": 1}, deletedFiles)
}

func TestExplainProcessor_ExecuteExplainRequest(t *testing.T) {
	t.Parallel()

	fileNames := []string{"first.go", "second.md"}

	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "/responses", r.URL.Path)

		var body map[string]any
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))

		input, ok := body["input"].(string)
		require.True(t, ok)
		require.Contains(t, input, fileNames[0])
		require.Contains(t, input, fileNames[1])

		require.Equal(t, explainInstruction, body["instructions"])

		tools, ok := body["tools"].([]any)
		require.True(t, ok)
		require.Len(t, tools, 1)

		tool, ok := tools[0].(map[string]any)
		require.True(t, ok)

		ids, ok := tool["vector_store_ids"].([]any)
		require.True(t, ok)
		require.Contains(t, ids, "vs_exec")

		w.Header().Set("Content-Type", "application/json")
		require.NoError(t, json.NewEncoder(w).Encode(newResponsePayload("generated explanation")))
	})

	ep := newTestExplainProcessor(client)
	ep.files = []*openai.FileObject{
		{ID: "file_a", Filename: fileNames[0]},
		{ID: "file_b", Filename: fileNames[1]},
	}
	ep.vectorStore = &openai.VectorStore{ID: "vs_exec"}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	result, err := ep.ExecuteExplainRequest(ctx)
	require.NoError(t, err)
	require.Equal(t, "generated explanation", result)
}

func newTestClient(t *testing.T, handler http.HandlerFunc) *openai.Client {
	t.Helper()

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	client := openai.NewClient(
		option.WithBaseURL(server.URL),
		option.WithAPIKey("test-key"),
	)

	return &client
}

func newVectorStore(id string, status openai.VectorStoreStatus, fileCount int) openai.VectorStore {
	now := time.Now().UTC().Unix()

	return openai.VectorStore{
		ID:        id,
		CreatedAt: now,
		FileCounts: openai.VectorStoreFileCounts{
			Cancelled:  0,
			Completed:  int64(fileCount),
			Failed:     0,
			InProgress: 0,
			Total:      int64(fileCount),
		},
		LastActiveAt: now,
		Metadata:     shared.Metadata{},
		Name:         "test-vector-store",
		Object:       "vector_store",
		Status:       status,
		UsageBytes:   0,
		ExpiresAfter: openai.VectorStoreExpiresAfter{
			Anchor: "last_active_at",
			Days:   1,
		},
	}
}

func newResponsePayload(text string) responses.Response {
	return responses.Response{
		ID:                "resp_1",
		CreatedAt:         float64(time.Now().UTC().Unix()),
		Error:             responses.ResponseError{Code: responses.ResponseErrorCodeServerError, Message: ""},
		IncompleteDetails: responses.ResponseIncompleteDetails{},
		Instructions: responses.ResponseInstructionsUnion{
			OfString: explainInstruction,
		},
		Metadata:          shared.Metadata{},
		Model:             explanationModel,
		Object:            "response",
		ParallelToolCalls: false,
		Temperature:       0,
		ToolChoice: responses.ResponseToolChoiceUnion{
			OfToolChoiceMode: responses.ToolChoiceOptionsAuto,
		},
		Tools: []responses.ToolUnion{},
		TopP:  1,
		Output: []responses.ResponseOutputItemUnion{
			{
				ID:     "msg_1",
				Type:   "message",
				Role:   "assistant",
				Status: string(responses.ResponseStatusCompleted),
				Content: []responses.ResponseOutputMessageContentUnion{
					{
						Annotations: []responses.ResponseOutputTextAnnotationUnion{},
						Text:        text,
						Type:        "output_text",
					},
				},
			},
		},
		Status: responses.ResponseStatusCompleted,
		Usage: responses.ResponseUsage{
			InputTokens: 1,
			InputTokensDetails: responses.ResponseUsageInputTokensDetails{
				CachedTokens: 0,
			},
			OutputTokens: 1,
			OutputTokensDetails: responses.ResponseUsageOutputTokensDetails{
				ReasoningTokens: 0,
			},
			TotalTokens: 2,
		},
	}
}

func TestRetry(t *testing.T) {
	t.Parallel()

	t.Run("succeeds after retries", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		callCount := 0
		val := "ready"

		got, err := retry(ctx, func() (*string, bool, error) {
			callCount++
			if callCount < 3 {
				return nil, false, nil
			}
			return &val, true, nil
		}, 0, 5)

		require.NoError(t, err)
		require.Equal(t, &val, got)
		require.Equal(t, 3, callCount)
	})

	t.Run("fails after too many attempts", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		callCount := 0

		got, err := retry(ctx, func() (*string, bool, error) {
			callCount++
			return nil, false, nil
		}, 0, 3)

		require.ErrorIs(t, err, ErrTooManyAttempts)
		require.Nil(t, got)
		require.Equal(t, 3, callCount)
	})

	t.Run("stops on context cancellation", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithCancel(context.Background())
		callCount := 0

		got, err := retry(ctx, func() (*string, bool, error) {
			callCount++
			cancel()
			return nil, false, nil
		}, 0, 5)

		require.ErrorIs(t, err, ErrContextCancelled)
		require.Nil(t, got)
		require.Equal(t, 1, callCount)
	})
}

func newTestExplainProcessor(client *openai.Client) *explainProcessor {
	return &explainProcessor{
		client: client,
		mutex:  &sync.RWMutex{},
	}
}
