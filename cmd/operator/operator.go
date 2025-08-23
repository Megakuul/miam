package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/pulumi/pulumi/sdk/v3/go/common/tokens"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/spf13/pflag"
)

type Flags struct {
	Config string
}

func ReadFlags() *Flags {
	flags := &Flags{}
	pflag.StringVarP(&flags.Config, "config", "c", "config.toml", "Specify a custom config file")
	pflag.Parse()
	return flags
}

type Config struct {
	Project string `toml:"project" env:"PROJECT" env-default:"miam"`
	Source  string `toml:"source" env:"SOURCE" env-default:"https://github.com/megakuul/miam"`
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		select {
		case <-sigs:
			cancel()
			return
		case <-ctx.Done():
			return
		}
	}()

	if err := run(ctx); err != nil {
		os.Stderr.WriteString("ERROR: " + err.Error())
		os.Exit(1)
	}
	return
}

func run(ctx context.Context) error {
	flags := ReadFlags()
	config := &Config{}
	if err := cleanenv.ReadConfig(flags.Config, config); err !=nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("cannot acquire file config: %v", err)
		}
	}
	if err := cleanenv.ReadEnv(config); err != nil {
		return fmt.Errorf("cannot acquire env config: %v", err)
	}
	err := tokens.ValidateProjectName(config.Project)
	if err != nil {
		return fmt.Errorf("invalid project name: %v", err)
	}

	return nil
}

func Deploy(ctx *pulumi.Context) error {
	return nil
}
