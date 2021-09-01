package config

import "github.com/hashicorp/hcl/v2"

type PageConfigList struct {
	Configs []PageConfig `hcl:"page,block"`
}

type PageConfig struct {
	Name    string           `hcl:"name,label"`
	Path    string           `hcl:"path,optional"`
	Hosts   []string         `hcl:"hosts,optional"`
	Enabled *bool            `hcl:"enabled,optional"`
	Source  PageSourceConfig `hcl:"source,block"`
}

type PageSourceConfig struct {
	Type    string   `hcl:"type,label"`
	Options hcl.Body `hcl:",remain"`
}
