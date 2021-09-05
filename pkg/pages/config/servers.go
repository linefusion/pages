package config

type ServerConfig struct {
	Name   string      `hcl:"name,label"`
	Listen ListenBlock `hcl:"listen,block"`
	Pages  PageBlocks  `hcl:"pages,block"`
}

type ListenBlock struct {
	Bind string `hcl:"bind,optional"`
	Port int    `hcl:"port"`
}
