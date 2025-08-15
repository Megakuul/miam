package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/iam"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/lambda"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optdestroy"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optup"
	"github.com/pulumi/pulumi/sdk/v3/go/common/tokens"
	"github.com/pulumi/pulumi/sdk/v3/go/common/workspace"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/spf13/pflag"
)

type Flags struct {
	Pocket bool
	Config string
}

func ReadFlags() *Flags {
	flags := &Flags{}
	pflag.BoolVarP(&flags.Pocket, "pocket-rocket", "p", false, "Bootstrap operator (this command can be used locally)")
	pflag.StringVarP(&flags.Config, "config", "c", "config.toml", "Specify a custom config file")
	pflag.Parse()
	return flags
}

type Config struct {
	Stack   string `toml:"stack" env:"STACK" env-default:"dev"`
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
	if err := tokens.ValidateProjectName(config.Project); err != nil {
		return fmt.Errorf("invalid project name: %v", err)
	}

	ws, err := auto.NewLocalWorkspace(ctx, auto.Project(workspace.Project{
		Name:    tokens.PackageName(config.Project),
		Runtime: workspace.NewProjectRuntimeInfo("go"),
		Backend: &workspace.ProjectBackend{
			URL: "s3://",
		},
	}))
	stack, err := auto.UpsertStackInlineSource(ctx, config.Stack, config.Project, Deploy)
	if err != nil {
		return fmt.Errorf("failed to construct stack: %v", err)
	}
	result, err := stack.Up(ctx)
	if err != nil {
		return fmt.Errorf("failed to update stack: %v", err)
	}

	some := result.Summary.Config[""]
}

func Deploy(ctx *pulumi.Context) error {

}
