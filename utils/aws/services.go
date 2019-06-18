package awsutils

import (
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/spf13/viper"
	"path"
)

func GetInvokeApiArn(restApiId string, stage string, method string, resource string) *arn.ARN {
	return &arn.ARN{
		Service: "execute-api",
		Region: viper.GetString("region"),
		AccountID: *GetIdentity().Account,
		Resource: path.Join(restApiId, stage, method, resource),
		Partition: "aws",
	}
}

func GetInvokeWsApiArn(wsApiId string, routeKey string) *arn.ARN {
	return &arn.ARN{
		Service: "execute-api",
		Region: viper.GetString("region"),
		AccountID: *GetIdentity().Account,
		Resource: path.Join(wsApiId, "*", routeKey),
		Partition: "aws",
	}
}

func GetAuthorizerArn(restApiId string, authorizerId string) *arn.ARN {
	return &arn.ARN{
		Service: "execute-api",
		Region: viper.GetString("region"),
		AccountID: *GetIdentity().Account,
		Resource: path.Join(restApiId, "authorizers", authorizerId),
		Partition: "aws",
	}
}

func GetGatewayLambdaInvokeArn(functionArn string) *arn.ARN {
	return &arn.ARN{
		Service: "apigateway",
		Region: viper.GetString("region"),
		AccountID: "lambda",
		Resource: path.Join("path", "2015-03-31", "functions", functionArn, "invocations"),
		Partition: "aws",
	}
}