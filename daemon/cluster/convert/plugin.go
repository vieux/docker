package convert

import (
	types "github.com/docker/docker/api/types/swarm"
	swarmapi "github.com/docker/swarmkit/api"
)

func pluginSpecFromGRPC(p *swarmapi.PluginSpec) *types.PluginSpec {
	if p == nil {
		return nil
	}
	pluginSpec := &types.PluginSpec{
		Image:   p.Image,
		Enabled: p.Enabled,
	}

	return pluginSpec
}

func pluginToGRPC(p *types.PluginSpec) (*swarmapi.PluginSpec, error) {
	if p == nil {
		return nil, nil
	}

	pluginSpec := &swarmapi.PluginSpec{
		Image:   p.Image,
		Enabled: p.Enabled,
	}

	return pluginSpec, nil
}
