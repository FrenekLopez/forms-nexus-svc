package main

import (
	"os"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsapigatewayv2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsapigatewayv2integrations"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsdynamodb"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type DeploymentsStackProps struct {
	awscdk.StackProps
}

func NewDeploymentsStack(scope constructs.Construct, id string, props *DeploymentsStackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}

	stack := awscdk.NewStack(scope, &id, &sprops)

	formsTable := awsdynamodb.NewTable(stack, jsii.String("FormsNexusTable"), &awsdynamodb.TableProps{
		PartitionKey: &awsdynamodb.Attribute{
			Name: jsii.String("id"),
			Type: awsdynamodb.AttributeType_STRING,
		},
		BillingMode:   awsdynamodb.BillingMode_PAY_PER_REQUEST,
		RemovalPolicy: awscdk.RemovalPolicy_DESTROY,
	})

	formsLambda := awslambda.NewFunction(stack, jsii.String("FormsNexusLambda"), &awslambda.FunctionProps{
		Architecture: awslambda.Architecture_ARM_64(),
		Runtime:      awslambda.Runtime_PROVIDED_AL2023(),
		Handler:      jsii.String("bootstrap"),
		Code:         awslambda.AssetCode_FromAsset(jsii.String("../bin"), nil),
		Timeout:      awscdk.Duration_Seconds(jsii.Number(30)),
		MemorySize:   jsii.Number(128),

		Environment: &map[string]*string{
			"DYNAMODB_TABLE_NAME": formsTable.TableName(),
			"SES_FROM_ADDRESS":    jsii.String(os.Getenv("SECRET_APP_EMAIL")),
			"SES_TO_ADDRESS":      jsii.String(os.Getenv("SECRET_APP_EMAIL")),
		},
	})

	sesPolicyStatement := awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
		Effect:    awsiam.Effect_ALLOW,
		Actions:   jsii.Strings("ses:SendEmail", "ses:SendRawEmail"),
		Resources: jsii.Strings("*"),
	})

	formsLambda.AddToRolePolicy(sesPolicyStatement)

	formsTable.Grant(formsLambda, jsii.String("dynamodb:PutItem"))

	lambdaIntegration := awsapigatewayv2integrations.NewHttpLambdaIntegration(
		jsii.String("FormsLambdaIntegration"),
		formsLambda,
		nil,
	)

	httpApi := awsapigatewayv2.NewHttpApi(stack, jsii.String("FormsNexusHttpApi"), &awsapigatewayv2.HttpApiProps{
		ApiName: jsii.String("FormsNexusInternalService"),
	})

	httpApi.AddRoutes(&awsapigatewayv2.AddRoutesOptions{
		Path:        jsii.String("/notifications"),
		Methods:     &[]awsapigatewayv2.HttpMethod{awsapigatewayv2.HttpMethod_POST},
		Integration: lambdaIntegration,
	})

	awscdk.NewCfnOutput(stack, jsii.String("ApiGatewayUrOutput"), &awscdk.CfnOutputProps{
		Value:       httpApi.Url(),
		Description: jsii.String("Public URL to send notifications via POST"),
	})

	awscdk.Tags_Of(formsLambda).Add(jsii.String("Environment"), jsii.String("Production"), nil)

	return stack
}

func main() {
	defer jsii.Close()

	app := awscdk.NewApp(nil)

	NewDeploymentsStack(app, "FormsNexusServiceStack", &DeploymentsStackProps{
		awscdk.StackProps{
			Env: &awscdk.Environment{
				Account: jsii.String(os.Getenv("CDK_DEFAULT_ACCOUNT")),
				Region:  jsii.String(os.Getenv("CDK_DEFAULT_REGION")),
			},
		},
	})

	app.Synth(nil)
}
