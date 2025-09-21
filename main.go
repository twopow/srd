package main

import (
	"fmt"

	"github.com/alecthomas/kong"
	kongyaml "github.com/alecthomas/kong-yaml"

	"srd/internal/cache"
	"srd/internal/log"
	"srd/internal/resolver"
	"srd/internal/server"
)

type Context struct {
	Debug bool
}

type ServeCmd struct {
	Server   server.ServerConfig     `embed:"" prefix:"server."`
	Log      log.LogConfig           `embed:"" prefix:"log."`
	Resolver resolver.ResolverConfig `embed:"" prefix:"resolver."`
	Cache    cache.CacheConfig       `embed:"" prefix:"cache."`
}

func (s *ServeCmd) Run(ctx *Context) error {
	log.NewLogger(ctx.Debug, s.Log.Pretty)
	cache := cache.New(s.Cache)
	resolver.Init(s.Resolver, cache)
	server.Start(s.Server)

	return nil
}

type CLI struct {
	Config kong.ConfigFlag `name:"config" type:"existingfile" help:"Path to config yaml file." env:"CONFIG_FILE"`

	Debug bool     `help:"Enable debug logging." env:"DEBUG"`
	Serve ServeCmd `cmd:"" help:"Run the HTTP server."`
}

func main() {
	var cli CLI
	ctx := kong.Parse(&cli,
		kong.Name("srd"),
		kong.Description("srd server"),
		kong.Configuration(kongyaml.Loader, "config.yaml", "config.yml", ".config.yaml", ".config.yml"),
		kong.DefaultEnvars("SRD"),
	)

	if cli.Debug {
		fmt.Printf("cli: %+v\n", cli)
	}

	err := ctx.Run(&Context{Debug: cli.Debug})
	ctx.FatalIfErrorf(err)
}
