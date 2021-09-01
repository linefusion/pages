package config

type ServerConfig struct {
	Name   string             `hcl:"name,label"`
	Listen ServerListenConfig `hcl:"listen,block"`
	Pages  PageConfigList     `hcl:"pages,block"`
}

type ServerListenConfig struct {
	Bind string `hcl:"bind,optional"`
	Port int    `hcl:"port"`
}
