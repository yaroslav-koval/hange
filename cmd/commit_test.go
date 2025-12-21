package cmd

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	aiagent_mock "github.com/yaroslav-koval/hange/mocks/aiagent"
	appbuilder_mock "github.com/yaroslav-koval/hange/mocks/appbuilder"
	changesprovider_mock "github.com/yaroslav-koval/hange/mocks/changesprovider"
	"github.com/yaroslav-koval/hange/pkg/agent/entity"
)

func TestCommitCmdRunESuccess(t *testing.T) {
	t.Parallel()

	gitMock := changesprovider_mock.NewMockChangesProvider(t)
	agentMock := aiagent_mock.NewMockAIAgent(t)

	app := appbuilder_mock.NewMockAppBuilder(t)

	app.EXPECT().GetGitChangesProvider().Return(gitMock, nil)
	app.EXPECT().GetAIAgent().Return(agentMock, nil)

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
