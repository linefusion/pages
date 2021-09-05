package config

import (
	"errors"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/joho/godotenv"
	"github.com/linefusion/pages/pkg/pages/config/funcs"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
)

type Config struct {
	Servers []ServerConfig `hcl:"server,block"`
}

func init() {
	envFile := os.Getenv("ENV_FILE")
	if envFile == "" {
		godotenv.Load()
	} else {
		godotenv.Load(envFile)
	}
}

func getEnv() map[string]cty.Value {

	env := map[string]cty.Value{}

	envPrefix, hasEnvPrefix := os.LookupEnv("LF_PAGES_PREFIX")
	if !hasEnvPrefix {
		envPrefix = "PAGES_"
	}

	for _, e := range os.Environ() {
		v := strings.SplitN(e, "=", 2)
		if len(v) != 2 {
			continue
		}
		key := v[0]
		value := v[1]
		if strings.HasPrefix(key, envPrefix) {
			key = strings.TrimPrefix(key, envPrefix)
			env[key] = cty.StringVal(value)
		}
	}

	return env
}

func GetDefaultVariables() map[string]cty.Value {

	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	hostname, err := os.Hostname()
	if err != nil {
		log.Fatal(err)
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	cacheDir, err := os.UserCacheDir()
	if err != nil {
		log.Fatal(err)
	}

	configDir, err := os.UserConfigDir()
	if err != nil {
		log.Fatal(err)
	}

	variables := map[string]cty.Value{}

	variables["process"] = cty.ObjectVal(map[string]cty.Value{
		"id": cty.NumberIntVal(int64(os.Getpid())),
	})

	variables["os"] = cty.ObjectVal(map[string]cty.Value{
		"hostname": cty.StringVal(hostname),
	})

	variables["dirs"] = cty.ObjectVal(map[string]cty.Value{
		"wd":     cty.StringVal(cwd),
		"cwd":    cty.StringVal(cwd),
		"pwd":    cty.StringVal(cwd),
		"config": cty.StringVal(configDir),
		"cache":  cty.StringVal(cacheDir),
		"home":   cty.StringVal(homeDir),
		"temp":   cty.StringVal(os.TempDir()),
		"tmp":    cty.StringVal(os.TempDir()),
	})

	env := getEnv()
	variables["env"] = cty.ObjectVal(env)

	return variables
}

func CreateEmptyContext() hcl.EvalContext {
	return hcl.EvalContext{
		Variables: map[string]cty.Value{},
		Functions: map[string]function.Function{},
	}
}

func CreateDefaultContext() hcl.EvalContext {
	context := CreateEmptyContext()
	context.Variables = GetDefaultVariables()
	context.Functions = funcs.AllFuncs
	return context
}

func CreateRequestContext(request *http.Request) hcl.EvalContext {
	context := CreateDefaultContext()

	params := map[string]cty.Value{}
	for paramKey, paramValues := range request.URL.Query() {
		values := []cty.Value{}
		for _, paramValue := range paramValues {
			values = append(values, cty.StringVal(paramValue))
		}
		params[paramKey] = cty.ListVal(values)
	}

	headers := map[string]cty.Value{}
	for headerKey, headerValues := range request.Header {
		values := []cty.Value{}
		for _, headerValue := range headerValues {
			values = append(values, cty.StringVal(headerValue))
		}
		headers[headerKey] = cty.ListVal(values)
	}

	context.Variables["request"] = cty.ObjectVal(map[string]cty.Value{
		"method":  cty.StringVal(request.Method),
		"scheme":  cty.StringVal(request.URL.Scheme),
		"host":    cty.StringVal(request.Host),
		"path":    cty.StringVal(request.URL.Path),
		"params":  cty.ObjectVal(params),
		"headers": cty.ObjectVal(headers),
	})

	return context
}

func load(parser *hclparse.Parser, file *hcl.File, diagnostics hcl.Diagnostics) (Config, error) {
	if diagnostics.HasErrors() {
		hcl.NewDiagnosticTextWriter(os.Stdout, parser.Files(), 78, true).WriteDiagnostics(diagnostics)
		log.Fatal("Error")
	}

	context := CreateDefaultContext()

	var config Config
	diagnostics = gohcl.DecodeBody(file.Body, &context, &config)
	if diagnostics.HasErrors() {
		hcl.NewDiagnosticTextWriter(os.Stdout, parser.Files(), 78, true).WriteDiagnostics(diagnostics)
		return config, errors.New(diagnostics.Error())
	}

	return config, nil
}

func LoadFile(path string, variables map[string]cty.Value) (Config, error) {
	var diagnostics hcl.Diagnostics
	parser := hclparse.NewParser()
	file, diagnostics := parser.ParseHCLFile(path)
	return load(parser, file, diagnostics)
}

func LoadString(str string, variables map[string]cty.Value) (Config, error) {
	var diagnostics hcl.Diagnostics
	parser := hclparse.NewParser()
	file, diagnostics := parser.ParseHCL([]byte(str), "Pagesfile")
	return load(parser, file, diagnostics)
}
