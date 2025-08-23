package main

import (
	"bufio"
	"context"
	"fmt"
	"strings"

	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optdestroy"
)

// nuke performs an interactive process to destroy a running operator.
func nuke(ctx context.Context, scanner *bufio.Scanner, ws auto.Workspace, stackName string) error {
	stack, err := auto.SelectStack(ctx, stackName, ws)
	if err != nil {
		return fmt.Errorf("failed to load stack: %v", err)
	}
	fmt.Println("🔸 Loading destruction preview")
	preview, err := stack.PreviewDestroy(ctx, optdestroy.Color("always"))
	if err != nil {
		return fmt.Errorf("stack dry run failed: %v", err)
	}
	fmt.Println()
	fmt.Println(preview.StdOut)
	if preview.StdErr != "" {
		fmt.Println()
		fmt.Println("⚠️ Anomalies detected in destruction preview")
	}
	fmt.Println()
	fmt.Println("🔹 Destroy the stack? [y/N]")
	scanner.Scan()
	if strings.ToUpper(scanner.Text()) == "y" {
		_, err = stack.Destroy(ctx)
		if err != nil {
			return fmt.Errorf("failed to destroy stack: %v", err)
		}
	} else {
		return fmt.Errorf("process cancelled")
	}
	return nil
}
