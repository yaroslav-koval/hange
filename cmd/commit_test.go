package cmd

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	aiagent_mock "github.com/yaroslav-koval/hange/mocks/aiagent"
	changesprovider_mock "github.com/yaroslav-koval/hange/mocks/changesprovider"
	"github.com/yaroslav-koval/hange/pkg/agent/entity"
	"github.com/yaroslav-koval/hange/pkg/factory"
)

func TestCommitCmdRunESuccess(t *testing.T) {
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
	gitMock.EXPECT().Commit(ctx, "final message").Return(nil)

	commitCmd.SetContext(ctx)

	err := commitCmd.RunE(commitCmd, nil)
	require.NoError(t, err)
}
