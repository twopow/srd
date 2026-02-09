package main

import (
	"fmt"
	"time"

	"github.com/alecthomas/kong"
	kongyaml "github.com/alecthomas/kong-yaml"

	"github.com/twopow/srd/internal/log"
	"github.com/twopow/srd/internal/server"
	"github.com/twopow/srd/resolver"
)

type Context struct {
	Debug bool
}

type ServeCmd struct {
	Server   server.ServerConfig `embed:"" prefix:"server."`
	Log      log.LogConfig       `embed:"" prefix:"log."`
	Resolver ResolverConfig      `embed:"" prefix:"resolver."`
}

type ResolverConfig struct {
	RecordPrefix       string `help:"Record prefix." default:"_srd"`
	NoHostBaseRedirect string `help:"No host base redirect." default:"https://github.com/twopow/srd"`

	TTL             time.Duration `help:"Cache TTL in seconds." default:"300s"`
	CleanupInterval time.Duration `help:"Cache cleanup interval in seconds." default:"900s"`
}

func (s *ServeCmd) Run(ctx *Context) error {
	log.NewLogger(ctx.Debug, s.Log.Pretty)

	rp, err := resolver.New(resolver.ResolverConfig{
		RecordPrefix:       s.Resolver.RecordPrefix,
		NoHostBaseRedirect: s.Resolver.NoHostBaseRedirect,
		TTL:                s.Resolver.TTL,
		CleanupInterval:    s.Resolver.CleanupInterval,
	})

	if err != nil {
		return fmt.Errorf("failed to init resolver: %w", err)
	}

	server.Start(s.Server, rp)

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
