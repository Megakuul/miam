package operator

import (
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/apigatewayv2"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/s3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Deploy(ctx *pulumi.Context) error {
	api, err := apigatewayv2.NewApi(ctx, "gateway", &apigatewayv2.ApiArgs{
		Name: pulumi.StringPtr(ctx.Project()),
		ProtocolType: pulumi.String("HTTP"),
	})
	if err!=nil {
		return err
	}
	integration, err := apigatewayv2.NewIntegration(ctx, "lambda", &apigatewayv2.IntegrationArgs{
		ApiId: api.ID(),
		ConnectionType: pulumi.String("INTERNET"),
		IntegrationType: pulumi.String("AWS_PROXY"),
		IntegrationMethod: pulumi.String("AWS_PROXY"), 
		IntegrationUri: pulumi.String("arn:TODO"),
		PayloadFormatVersion: pulumi.String("2.0"),
	})
	if err!=nil {
		return err
	}
	_, err = apigatewayv2.NewRoute(ctx, "route", &apigatewayv2.RouteArgs{
		ApiId: api.ID(),
		OperationName: pulumi.String("deploy"),
		RouteKey: pulumi.String("/deploy"),
		Target: pulumi.Sprintf("integrations/%s", integration.ID()),
	})
	if err!=nil {
		return err
	}
	return nil
}
