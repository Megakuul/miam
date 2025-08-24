package main

import (
	"context"
	"fmt"

	"github.com/pterm/pterm"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optpreview"
)

// launch performs an interactive process to deploy the operator stack on the provided workspace.
func launch(ctx context.Context, ws auto.Workspace) error {
	environment, _ := pterm.DefaultInteractiveTextInput.
		WithDefaultValue("prod").Show("Enter the environment")
	spinner, _ := pterm.DefaultSpinner.WithRemoveWhenDone(true).
		Start("Merging stack update locally...")
	defer spinner.Stop()
	stack, err := auto.UpsertStack(ctx, environment, ws)
	if err != nil {
		return fmt.Errorf("failed to construct stack: %v", err)
	}
	spinner.Stop()
	spinner, _ = pterm.DefaultSpinner.WithRemoveWhenDone(true).
		Start("Loading stack update preview...")
	preview, err := stack.Preview(ctx, optpreview.Color("always"))
	if err != nil {
		return fmt.Errorf("stack preview failed: %v", err)
	}
	spinner.Stop()
	fmt.Println()
	fmt.Println(preview.StdOut)
	if preview.StdErr != "" {
		fmt.Println()
		fmt.Println("⚠️ Anomalies detected in deployment preview")
	}
	fmt.Println()
	ok, _ := pterm.DefaultInteractiveConfirm.
		WithDefaultValue(false).Show("Deploy the operator?")
	if !ok {
		return fmt.Errorf("process cancelled")
	}
	spinner, _ = pterm.DefaultSpinner.WithRemoveWhenDone(true).
		Start("Applying stack update...")
	_, err = stack.Up(ctx)
	if err != nil {
		return fmt.Errorf("failed to update stack: %v", err)
	}
	return nil
}
