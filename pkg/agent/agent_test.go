package agent

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	explainprocessor_mock "github.com/yaroslav-koval/hange/mocks/explainprocessor"
	"github.com/yaroslav-koval/hange/pkg/entities"
)

func TestExplainFilesSuccess(t *testing.T) {
	ep := explainprocessor_mock.NewMockExplainProcessor(t)

	files := make(chan entities.File)
	close(files)

	ep.EXPECT().UploadFiles(mock.Anything, mock.Anything).Return(nil)
	ep.EXPECT().ExecuteExplainRequest(mock.Anything).Return("ok", nil)
	ep.EXPECT().Cleanup(mock.Anything)

	uc := &agent{ep: ep}

	result, err := uc.ExplainFiles(context.Background(), files)
	require.NoError(t, err)
	require.Equal(t, "ok", result)
}

func TestExplainFilesUploadFails(t *testing.T) {
	ep := explainprocessor_mock.NewMockExplainProcessor(t)

	files := make(chan entities.File)
	close(files)

	uploadErr := errors.New("upload failed")

	ep.EXPECT().UploadFiles(mock.Anything, mock.Anything).Return(uploadErr)
	ep.EXPECT().Cleanup(mock.Anything)

	uc := &agent{ep: ep}

	result, err := uc.ExplainFiles(context.Background(), files)
	require.ErrorIs(t, err, uploadErr)
	require.Empty(t, result)
	ep.AssertNotCalled(t, "ExecuteExplainRequest", mock.Anything)
}
