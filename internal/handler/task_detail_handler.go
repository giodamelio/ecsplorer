package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/masaushi/ecsplorer/internal/api"
	"github.com/masaushi/ecsplorer/internal/app"
	"github.com/masaushi/ecsplorer/internal/view"
)

func TaskDetailHandler(ctx context.Context, operator app.Operator) (app.Page, error) {
	cluster := valueFromContext[*types.Cluster](ctx)
	task := valueFromContext[*types.Task](ctx)

	return view.NewTaskDetail(task).
		SetContainerSelectAction(func(container *types.Container) {
			operator.ConfirmModal("Exec shell against the container?", func() {
				// TODO: refactor
				execSess, err := operator.ECS().CreateExecuteSession(ctx, &api.ECSCreateExecuteSessionParams{
					Cluster:   cluster,
					Task:      task,
					Container: container,
					Command:   "/bin/sh",
				})
				if err != nil {
					operator.ErrorModal(err)
					return
				}

				sess, err := json.Marshal(execSess)
				if err != nil {
					operator.ErrorModal(err)
					return
				}

				target := fmt.Sprintf("ecs:%s_%s_%s",
					aws.ToString(cluster.ClusterArn),
					aws.ToString(task.TaskArn),
					aws.ToString(container.RuntimeId),
				)
				ssmTarget, err := json.Marshal(map[string]string{"Target": target})
				if err != nil {
					operator.ErrorModal(err)
					return
				}

				operator.Suspend(func() {
					//nolint:gosec
					cmd := exec.Command(
						"session-manager-plugin",
						string(sess),
						operator.Region(),
						"StartSession",
						"",
						string(ssmTarget),
						"https://ssm.ap-northeast-1.amazonaws.com",
					)
					cmd.Stdout = os.Stdout
					cmd.Stdin = os.Stdin
					cmd.Stderr = os.Stderr
					err = cmd.Run()
				})

				if err != nil {
					operator.ErrorModal(err)
				}
			})
		}).
		SetPrevPageAction(func() {
			operator.Goto(ctx, ClusterDetailHandler)
		}), nil
}
