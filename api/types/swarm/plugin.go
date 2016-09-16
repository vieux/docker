package swarm

// PluginSpec represents the spec of a plugin.
type PluginSpec struct {
	Image   string `json:",omitempty"`
	Enabled bool   `json:",omitempty"`
}
