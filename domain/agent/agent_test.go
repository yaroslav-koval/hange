package agent

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/yaroslav-koval/hange/domain/agent/entity"
	"github.com/yaroslav-koval/hange/domain/entities"
	commitprocessor_mock "github.com/yaroslav-koval/hange/mocks/commitprocessor"
	explainprocessor_mock "github.com/yaroslav-koval/hange/mocks/explainprocessor"
)

func TestExplainFilesSuccess(t *testing.T) {
	ep := explainprocessor_mock.NewMockExplainProcessor(t)

	files := make(chan entities.File)
	close(files)

	ep.EXPECT().ProcessFiles(mock.Anything, mock.Anything).Return(nil)
	ep.EXPECT().ExecuteExplainRequest(mock.Anything).Return("ok", nil)
	ep.EXPECT().Cleanup(mock.Anything)

	uc := newTestAgent(nil, ep)

	result, err := uc.ExplainFiles(context.Background(), files)
	require.NoError(t, err)
	require.Equal(t, "ok", result)
}

func TestExplainFilesUploadFails(t *testing.T) {
	ep := explainprocessor_mock.NewMockExplainProcessor(t)

	files := make(chan entities.File)
	close(files)

	uploadErr := errors.New("upload failed")

	ep.EXPECT().ProcessFiles(mock.Anything, mock.Anything).Return(uploadErr)
	ep.EXPECT().Cleanup(mock.Anything)

	uc := newTestAgent(nil, ep)

	result, err := uc.ExplainFiles(context.Background(), files)
	require.ErrorIs(t, err, uploadErr)
	require.Empty(t, result)
	ep.AssertNotCalled(t, "ExecuteExplainRequest", mock.Anything)
}

func TestCreateCommitMessage(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		cp := commitprocessor_mock.NewMockCommitProcessor(t)
		data := entity.CommitData{
			UserInput:    "task context",
			Status:       "status output",
			StagedStatus: "staged files",
			Diff:         "diff content",
		}

		cp.EXPECT().GenCommitMessage(mock.Anything, data).Return("commit message", nil)

		result, err := newTestAgent(cp, nil).CreateCommitMessage(context.Background(), data)
		require.NoError(t, err)
		require.Equal(t, "commit message", result)
	})

	t.Run("fails validation when statuses missing", func(t *testing.T) {
		cp := commitprocessor_mock.NewMockCommitProcessor(t)
		data := entity.CommitData{Diff: "diff content"}

		result, err := newTestAgent(cp, nil).CreateCommitMessage(context.Background(), data)
		require.ErrorIs(t, err, ErrNoStatusProvided)
		require.Empty(t, result)
		cp.AssertNotCalled(t, "GenCommitMessage", mock.Anything, mock.Anything)
	})

	t.Run("fails validation when diff missing", func(t *testing.T) {
		cp := commitprocessor_mock.NewMockCommitProcessor(t)
		data := entity.CommitData{
			Status:       "status output",
			StagedStatus: "staged files",
		}

		result, err := newTestAgent(cp, nil).CreateCommitMessage(context.Background(), data)
		require.ErrorIs(t, err, ErrProvidedEmptyInput)
		require.Empty(t, result)
		cp.AssertNotCalled(t, "GenCommitMessage", mock.Anything, mock.Anything)
	})
}

func TestValidateCommitParams(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		data entity.CommitData
		err  error
	}{
		{
			name: "no statuses",
			data: entity.CommitData{Diff: "diff content"},
			err:  ErrNoStatusProvided,
		},
		{
			name: "missing diff with statuses",
			data: entity.CommitData{Status: "status", StagedStatus: "staged"},
			err:  ErrProvidedEmptyInput,
		},
		{
			name: "missing status only",
			data: entity.CommitData{StagedStatus: "staged", Diff: "diff content"},
		},
		{
			name: "missing staged status only",
			data: entity.CommitData{Status: "status", Diff: "diff content"},
		},
		{
			name: "all provided",
			data: entity.CommitData{Status: "status", StagedStatus: "staged", Diff: "diff content"},
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := validateCommitParamsForTest(tt.data)
			if tt.err == nil {
				require.NoError(t, err)
				return
			}

			require.ErrorIs(t, err, tt.err)
		})
	}
}

// newTestAgent constructs an agent with provided collaborators for testing.
func newTestAgent(cp CommitProcessor, ep ExplainProcessor) *agent {
	return &agent{cp: cp, ep: ep}
}

// validateCommitParamsForTest exposes validation logic for external tests.
func validateCommitParamsForTest(data entity.CommitData) error {
	return (&agent{}).validateCommitParams(data)
}
