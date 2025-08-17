// package pocketrocket provides a bootstrap setup tui for the operator.
// It provides a simple launcher that solves the initial state bucket chicken egg problem.
package pocketrocket

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/common/tokens"
	"github.com/pulumi/pulumi/sdk/v3/go/common/workspace"
	"golang.org/x/term"

	"github.com/manifoldco/promptui"
)

// Setup performs an interactive process that acquires
// the storage backend for the pulumi stack.
func Setup(ctx context.Context) (auto.Workspace, error) {
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		width = 50
	}
	fmt.Println(generateHeader(width))

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("🔸 Bootstrapping aws client...")

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}

	fmt.Printf("🔹 Enter the project name: ")
	scanner.Scan()
	project := scanner.Text()

	bucket, prefix, err := setupBucket(ctx, cfg, scanner)
	if err != nil {
		return nil, fmt.Errorf("failed to setup bucket: %v", err)
	}
	_ = fmt.Sprintf("s3://%s.s3.amazonaws.com/%s", bucket, prefix)

	ws, err := auto.NewLocalWorkspace(ctx, auto.Project(workspace.Project{
		Name:    tokens.PackageName(project),
		Runtime: workspace.NewProjectRuntimeInfo("go", map[string]interface{}{}),
		Backend: &workspace.ProjectBackend{
			URL: "s3://",
		},
	}))
	if err != nil {
		return nil, err
	}
	_ = ws
	return nil, nil
}

func generateHeader(width int) string {
	if width < 100 {
		return "🚀 Welcome to the pocketrocket bootstrap process"
	}
	return `
                                      ▓██                                    
                                     ▓▓███                                   
                                    ▓▒█████                                  
                                   ▓▒▒▓▓██▓▓                                 
                                   ▒▒▒██████                                 
                                  ▒▒▒███████▓                                
                                  ▒▒████████▓                                
                                  ▒▒▒▓██████▓                                
                                  ▒▒▒▓▓█████▓                                
                                  ▒▒▒▓▓▓████▓                                
                                  ▒▒▒▓▓▓████▓                                
                                 ▒█▓▓▓▓▓█████▓                               
                                ▓██▓▓▓▓██████▓▓                              
                               ▓███▓▓▓▓██████▓▒▓                             
                               █▓▓ █▓███████  ▒▒                             
                                  ▒▒▒▒▒▓▒▒▓▒▒                                
                                 ▒▒▒▒▒▒▒▒▒▒▒▒▒                               
                                 ▒▒▒▒▒▒▒▒▒▒▒▒▒                               
                                ▒▒▒▒▒▒▒▒▒▒▒▓▒▒                               
                             ▒▒▒▒▒▒▒▒▒▒▒▒▒▒▓▓▓▓▓  ▒▒▓▓                       
                   ▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▓▓▓████▓▓▓▓▓▓▓                    
              ▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▓▓▓▒▒▒▒▒▒▒▒▓▓▓▓▓▓▓▓▓▓▓▒▓▓▓▓   ▓▒▒▒▒          
          ▒▒▒▒▒▓▒▒▓▓▓▒▒▒▒▒▒▒▓▓▓▓▓▓▓▒▒▒▒▒▒▒▒▓▓▒▓██▓▓▓▓▓▓▓▓▓█▓▓▓▓▒▒▓▓▒         
        ▒▒▒▒▒▒▓▓▒▒▓▒▓▒▒▒▒▒▒▓▓▓▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▓███▓▓▒▒▓▓█████▓▓▓▓▓▓▓▓       
   ▒▒▓▒▒▒▒▒▒▒▒▒▓▓▓▓▓▒▒▒▓▒▒▓▒▒▒▒▒▒▓▒▒▒▒▓▒▒▒▒▒▒▓▓▓▓▓▓▓▒▓▓▒▓▓▓▓▓▓▓▓▓▓█████▓▓▓▓  
  ▒▒▓▒▓▓▒▒▒▒▓▓▒▒▒▒▓▒▒▒▓▓▓▓▒▒▒▓▒▒▒▓▓▓▒▓▓▒▒▓▒▓▓▓▓█████▓▓▒▒▓▓█▓▓▓▓▓█████▓█▓█▓███
           █████▓████████████▓▒▓▓█████▓▓█████▓▓█████████████████             
		
           🚀 Welcome to the pocketrocket bootstrap process 🚀
		`
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
