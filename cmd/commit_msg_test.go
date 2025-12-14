package cmd

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/yaroslav-koval/hange/pkg/agent/entity"

	aiagent_mock "github.com/yaroslav-koval/hange/mocks/aiagent"
	changesprovider_mock "github.com/yaroslav-koval/hange/mocks/changesprovider"
	"github.com/yaroslav-koval/hange/pkg/factory"
)

func TestCommitMessageCommandRunESuccess(t *testing.T) {
	t.Parallel()

	gitMock := changesprovider_mock.NewMockChangesProvider(t)
	agentMock := aiagent_mock.NewMockAIAgent(t)

	app := &factory.App{
		Agent: agentMock,
		Git:   gitMock,
	}

	ctx := appToContext(context.Background(), app)

	gitMock.EXPECT().Status(ctx).Return("git status", nil)
	gitMock.EXPECT().StagedStatus(ctx).Return("staged status", nil)
	gitMock.EXPECT().StagedDiff(ctx, 30).Return("diff output", nil)
	agentMock.EXPECT().CreateCommitMessage(ctx, entity.CommitData{
		Status:       "git status",
		StagedStatus: "staged status",
		Diff:         "diff output",
	}).Return("final message", nil)

	commitMsgCmd.SetContext(ctx)

	err := commitMsgCmd.RunE(commitMsgCmd, nil)
	require.NoError(t, err)
}

func TestGenerateCommitMessageRejectsMultipleArgs(t *testing.T) {
	t.Parallel()

	message, err := generateCommitMessage(context.Background(), nil, []string{"one", "two"})
	require.Empty(t, message)
	require.ErrorContains(t, err, "at most 1 arg")
}

func TestGenerateCommitMessageSuccess(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		args      []string
		userInput string
	}{
		{
			name:      "without user input",
			args:      nil,
			userInput: "",
		},
		{
			name:      "with user input",
			args:      []string{"feature description"},
			userInput: "feature description",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			gitMock := changesprovider_mock.NewMockChangesProvider(t)
			agentMock := aiagent_mock.NewMockAIAgent(t)

			gitMock.EXPECT().Status(ctx).Return("git status", nil)
			gitMock.EXPECT().StagedStatus(ctx).Return("staged status", nil)
			gitMock.EXPECT().StagedDiff(ctx, 30).Return("diff output", nil)

			agentMock.EXPECT().CreateCommitMessage(ctx, entity.CommitData{
				UserInput:    tt.userInput,
				Status:       "git status",
				StagedStatus: "staged status",
				Diff:         "diff output",
			}).Return("final message", nil)

			app := &factory.App{
				Agent: agentMock,
				Git:   gitMock,
			}

			message, err := generateCommitMessage(ctx, app, tt.args)
			require.NoError(t, err)
			require.Equal(t, "final message", message)
		})
	}
}

func TestGenerateCommitMessagePropagatesErrors(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("status error", func(t *testing.T) {
		t.Parallel()

		statusErr := errors.New("status failed")

		gitMock := changesprovider_mock.NewMockChangesProvider(t)
		gitMock.EXPECT().Status(ctx).Return("", statusErr)

		app := &factory.App{Git: gitMock}

		message, err := generateCommitMessage(ctx, app, nil)
		require.Empty(t, message)
		require.ErrorIs(t, err, statusErr)
	})

	t.Run("staged status error", func(t *testing.T) {
		t.Parallel()

		stagedStatusErr := errors.New("staged status failed")

		gitMock := changesprovider_mock.NewMockChangesProvider(t)
		gitMock.EXPECT().Status(ctx).Return("git status", nil)
		gitMock.EXPECT().StagedStatus(ctx).Return("", stagedStatusErr)

		app := &factory.App{Git: gitMock}

		message, err := generateCommitMessage(ctx, app, nil)
		require.Empty(t, message)
		require.ErrorIs(t, err, stagedStatusErr)
	})

	t.Run("staged diff error", func(t *testing.T) {
		t.Parallel()

		diffErr := errors.New("staged diff failed")

		gitMock := changesprovider_mock.NewMockChangesProvider(t)
		gitMock.EXPECT().Status(ctx).Return("git status", nil)
		gitMock.EXPECT().StagedStatus(ctx).Return("staged status", nil)
		gitMock.EXPECT().StagedDiff(ctx, 30).Return("", diffErr)

		app := &factory.App{Git: gitMock}

		message, err := generateCommitMessage(ctx, app, nil)
		require.Empty(t, message)
		require.ErrorIs(t, err, diffErr)
	})

	t.Run("agent error", func(t *testing.T) {
		t.Parallel()

		agentErr := errors.New("agent failed")

		gitMock := changesprovider_mock.NewMockChangesProvider(t)
		gitMock.EXPECT().Status(ctx).Return("git status", nil)
		gitMock.EXPECT().StagedStatus(ctx).Return("staged status", nil)
		gitMock.EXPECT().StagedDiff(ctx, 30).Return("diff output", nil)

		agentMock := aiagent_mock.NewMockAIAgent(t)
		agentMock.EXPECT().CreateCommitMessage(ctx, entity.CommitData{
			Status:       "git status",
			StagedStatus: "staged status",
			Diff:         "diff output",
		}).Return("", agentErr)

		app := &factory.App{
			Agent: agentMock,
			Git:   gitMock,
		}

		message, err := generateCommitMessage(ctx, app, nil)
		require.Empty(t, message)
		require.ErrorIs(t, err, agentErr)
	})
}
