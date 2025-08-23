package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/common/tokens"
	"github.com/pulumi/pulumi/sdk/v3/go/common/workspace"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		select {
		case <-sigs:
			cancel()
			os.Exit(1)
		case <-ctx.Done():
			return
		}
	}()

	if err := run(ctx); err != nil {
		fmt.Println("❌ ========= ERROR =========")
		fmt.Println()
		color.RGB(255, 82, 82).Println(err.Error())
		fmt.Println("❌ ========= ERROR =========")
		os.Exit(1)
	}
	return
}

func run(ctx context.Context) error {
	fmt.Println(generateHeader())

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("🔸 Bootstrapping aws client...")
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return err
	}

	fmt.Printf("🔹 Enter the project name: ")
	scanner.Scan()
	project := scanner.Text()

	bucket, prefix, err := setupBucket(ctx, cfg, scanner)
	if err != nil {
		return fmt.Errorf("failed to setup bucket: %v", err)
	}

	secretArn, err := setupSecret(ctx, cfg, scanner)
	if err != nil {
		return fmt.Errorf("failed to setup secret: %v", err)
	}

	ws, err := auto.NewLocalWorkspace(ctx, auto.Project(workspace.Project{
		Name:    tokens.PackageName(project),
		Author:  aws.String("miam pocketrocket cli"),
		Runtime: workspace.NewProjectRuntimeInfo("go", map[string]any{}),
		Backend: &workspace.ProjectBackend{
			URL: fmt.Sprintf("s3://%s/%s", bucket, prefix),
		},
	}), auto.SecretsProvider(fmt.Sprintf("awskms://%s", secretArn)))
	if err != nil {
		return err
	}
	stacks, err := ws.ListStacks(ctx)
	if err != nil {
		return err
	}
	if len(stacks) < 1 {
		return launch(ctx, scanner, ws)
	} else {
		fmt.Printf("🔹 Enter action: [Launch/Nuke] ")
		scanner.Scan()
		switch strings.ToLower(scanner.Text()) {
		case "l", "launch":
			return launch(ctx, scanner, ws)
		case "n", "nuke":
			for _, stack := range stacks {
				err = nuke(ctx, scanner, ws, stack.Name)
				if err != nil {
					return err
				}
			}
			return nil
		default:
			return fmt.Errorf("not a valid action")
		}
	}
}

func setupSecret(ctx context.Context, cfg aws.Config, scanner *bufio.Scanner) (string, error) {
	smClient := secretsmanager.NewFromConfig(cfg)
	fmt.Printf("🔹 Use existing secret for infra state? [y/N] ")
	scanner.Scan()
	if strings.ToLower(scanner.Text()) == "y" {
		listResp, err := smClient.ListSecrets(ctx, &secretsmanager.ListSecretsInput{})
		if err != nil {
			return "", err
		}
		secrets := []string{}
		for _, secret := range listResp.SecretList {
			secrets = append(secrets, *secret.ARN)
		}
		prompt := promptui.Select{
			Label: "Select secret",
			Items: secrets,
		}
		_, selected, err := prompt.Run()
		if err != nil {
			return "", err
		}
		return selected, nil
	} else {
		fmt.Printf("🔹 Enter the secret name: ")
		scanner.Scan()
		name := scanner.Text()
		createResp, err := smClient.CreateSecret(ctx, &secretsmanager.CreateSecretInput{
			Name: aws.String(name),
			Description: aws.String("Secret used to encrypt sensitive pulumi stack data"),
		})
		if err != nil {
			return "", err
		}
		return *createResp.ARN, nil
	}
}

func setupBucket(ctx context.Context, cfg aws.Config, scanner *bufio.Scanner) (string, string, error) {
	s3Client := s3.NewFromConfig(cfg)
	fmt.Printf("🔹 Use existing s3 bucket for infra state? [y/N] ")
	scanner.Scan()
	if strings.ToLower(scanner.Text()) == "y" {
		listResp, err := s3Client.ListBuckets(ctx, &s3.ListBucketsInput{})
		if err != nil {
			return "", "", err
		}
		buckets := []string{}
		for _, bucket := range listResp.Buckets {
			buckets = append(buckets, *bucket.Name)
		}
		prompt := promptui.Select{
			Label: "Select bucket",
			Items: buckets,
		}
		_, selected, err := prompt.Run()
		if err != nil {
			return "", "", err
		}
		fmt.Printf("🔹 Specify bucket prefix: ")
		scanner.Scan()
		return selected, scanner.Text(), nil
	} else {
		fmt.Printf("🔹 Enter the bucket name: ")
		scanner.Scan()
		name := scanner.Text()
		fmt.Printf("🔹 Enter the bucket region [eu-central-1]: ")
		scanner.Scan()
		region := types.BucketLocationConstraint(scanner.Text())
		if region == "" {
			region = types.BucketLocationConstraintEuCentral1
		}
		createResp, err := s3Client.CreateBucket(ctx, &s3.CreateBucketInput{
			Bucket: aws.String(name),
			CreateBucketConfiguration: &types.CreateBucketConfiguration{
				LocationConstraint: region,
			},
		})
		if err != nil {
			return "", "", err
		}
		return strings.TrimPrefix(*createResp.Location, "/"), "", nil
	}
}
