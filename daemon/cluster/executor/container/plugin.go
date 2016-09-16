package container

import (
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/plugin"
	"github.com/docker/swarmkit/api"
	"golang.org/x/net/context"
)

// pluginController implements agent.Controller against docker's API.
//
// pluginController manages the lifecycle of a docker plugin.
type pluginController struct {
	manager *plugin.Manager
	task    *api.Task
}

func newPluginController(task *api.Task) (*pluginController, error) {
	return &pluginController{
		manager: plugin.GetManager(),
		task:    task,
	}, nil
}

func (pc *pluginController) Update(ctx context.Context, t *api.Task) error {
	if pc.task.Spec.GetPlugin().Enabled {
		return pc.manager.Enable(pc.task.Spec.GetPlugin().Image)
	}
	return pc.manager.Disable(pc.task.Spec.GetPlugin().Image)
}

func (pc *pluginController) Prepare(ctx context.Context) error {
	_, err := pc.manager.Pull(pc.task.Spec.GetPlugin().Image, nil, &types.AuthConfig{})
	return err
}

func (pc *pluginController) Start(ctx context.Context) error {
	if pc.task.Spec.GetPlugin().Enabled {
		return pc.manager.Enable(pc.task.Spec.GetPlugin().Image)
	}
	return nil
}

func (pc *pluginController) Wait(ctx context.Context) error {
	<-ctx.Done()
	return ctx.Err()
}

func (pc *pluginController) Shutdown(ctx context.Context) error {
	return pc.manager.Disable(pc.task.Spec.GetPlugin().Image)
}

func (pc *pluginController) Terminate(ctx context.Context) error {
	return pc.manager.Disable(pc.task.Spec.GetPlugin().Image)
}

func (pc *pluginController) Remove(ctx context.Context) error {
	return pc.manager.Remove(pc.task.Spec.GetPlugin().Image, &types.PluginRmConfig{ForceRemove: true})
}

func (pc *pluginController) Close() error {
	return nil
}
