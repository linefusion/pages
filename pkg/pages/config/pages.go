package config

import "github.com/hashicorp/hcl/v2"

type PageBlocks struct {
	Entries []PageBlock `hcl:"page,block"`
}

type PageBlock struct {
	Name    string      `hcl:"name,label"`
	Path    string      `hcl:"path,optional"`
	Hosts   []string    `hcl:"hosts,optional"`
	Enabled *bool       `hcl:"enabled,optional"`
	Source  SourceBlock `hcl:"source,block"`
}

type SourceBlock struct {
	Type    string   `hcl:"type,label"`
	Options hcl.Body `hcl:",remain"`
}
