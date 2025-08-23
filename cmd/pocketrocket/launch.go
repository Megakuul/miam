package main

import (
	"bufio"
	"context"
	"fmt"
	"strings"

	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optpreview"
)

// launch performs an interactive process to deploy the operator stack on the provided workspace.
func launch(ctx context.Context, scanner *bufio.Scanner, ws auto.Workspace) error {
	fmt.Printf("🔹 Enter the environment: [prod] ")
	scanner.Scan()
	environment := scanner.Text()
	if environment == "" {
		environment = "prod"
	}
	stack, err := auto.UpsertStack(ctx, environment, ws)
	if err != nil {
		return fmt.Errorf("failed to construct stack: %v", err)
	}
	fmt.Println("🔸 Loading deployment preview")
	preview, err := stack.Preview(ctx, optpreview.Color("always"))
	if err != nil {
		return fmt.Errorf("stack dry run failed: %v", err)
	}
	fmt.Println()
	fmt.Println(preview.StdOut)
	if preview.StdErr != "" {
		fmt.Println()
		fmt.Println("⚠️ Anomalies detected in deployment preview")
	}
	fmt.Println()
	fmt.Println("🔹 Deploy the operator? [y/N]")
	scanner.Scan()
	if strings.ToUpper(scanner.Text()) == "y" {
		_, err = stack.Up(ctx)
		if err != nil {
			return fmt.Errorf("failed to update stack: %v", err)
		}
	} else {
		return fmt.Errorf("process cancelled")
	}
	return nil
}
