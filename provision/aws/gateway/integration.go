package gateway

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/niranjan94/bifrost/config"
	"github.com/niranjan94/bifrost/provision/aws/functions"
	"github.com/niranjan94/bifrost/utils"
	awsutils "github.com/niranjan94/bifrost/utils/aws"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"path"
	"strings"
)

func addInvokePermission(aliasArn string, invokeArn string) error {
	lambdaSvc := lambda.New(awsutils.GetSession())
	statementId := utils.SHA1Hash(invokeArn)

	_, _ = lambdaSvc.RemovePermission(&lambda.RemovePermissionInput{
		FunctionName: &aliasArn,
		StatementId:  &statementId,
	})
	_, err := lambdaSvc.AddPermission(&lambda.AddPermissionInput{
		FunctionName: &aliasArn,
		Principal:    aws.String("apigateway.amazonaws.com"),
		Action:       aws.String("lambda:InvokeFunction"),
		StatementId:  &statementId,
		SourceArn:    &invokeArn,
	})
	return err
}

func IntegrateFunctions(functions []*functions.DeploymentPackage) error {

	restApiId := config.GetString("apiGateway.restApiId")
	if restApiId == "" {
		return nil
	}

	logrus.Info("querying API Gateway")

	gatewaySvc := apigateway.New(awsutils.GetSession())

	_, err := gatewaySvc.GetRestApi(&apigateway.GetRestApiInput{
		RestApiId: &restApiId,
	})
	if err != nil {
		return err
	}

	var resources []*apigateway.Resource

	if err = gatewaySvc.GetResourcesPages(
		&apigateway.GetResourcesInput{
			Limit: aws.Int64(500),
			RestApiId: &restApiId,
		},
		func(output *apigateway.GetResourcesOutput, b bool) bool {
			resources = append(resources, output.Items...)
			return true
		},
	); err != nil {
		return err
	}

	resourcePrefix := viper.GetString("apiGateway.resourcePrefix")

	getResourceByPath := func(resourcePath string) *apigateway.Resource {
		fullResourcePath := path.Join(resourcePrefix, resourcePath)
		if strings.HasPrefix(resourcePath, "/") {
			fullResourcePath = resourcePath
		}
		for idx := range resources {
			resource := resources[idx]
			if *resource.Path == fullResourcePath {
				return resource
			}
		}
		return nil
	}

	for _, function := range functions {
		cfg := function.Config
		stage := cfg.GetString("stage")

		logrus.Infof("activating %s on REST API %s", function.FunctionName, restApiId)

		gatewayLambdaInvocationArn := awsutils.GetGatewayLambdaInvokeArn(function.FunctionArn + ":${stageVariables.lambdaAlias}").String()

		resources := cfg.GetStringSlice("api.resources")
		if singleResource := cfg.GetString("api.resource"); singleResource != "" {
			resources = append(resources, singleResource)
		}

		for _, resourceString := range resources {
			if resource := strings.Split(cfg.GetString(resourceString), ":"); len(resource) >= 2 {
				method := strings.ToUpper(resource[0])
				resourcePath := resource[1]
				invokeArn := awsutils.GetInvokeApiArn(restApiId, stage, method, path.Join(resourcePrefix, resourcePath)).String()
				resource := getResourceByPath(resourcePath)
				if resource != nil {
					logrus.Infof("updating integration for %s resource %s", method, resourcePath)
					if _, err := gatewaySvc.UpdateIntegration(&apigateway.UpdateIntegrationInput{
						RestApiId: &restApiId,
						ResourceId: resource.Id,
						HttpMethod: &method,
						PatchOperations: []*apigateway.PatchOperation{
							{
								Op:    aws.String(apigateway.OpReplace),
								Path:  aws.String("/uri"),
								Value: &gatewayLambdaInvocationArn,
							},
						},
					}); err != nil {
						logrus.Error(err)
					}

					if err := addInvokePermission(function.AliasArn, invokeArn); err != nil {
						logrus.Error(err)
						continue
					}
				} else {
					logrus.Errorf("could not find API Gateway resource %s for %s", resourcePath, function.FunctionName)
				}
			}
		}

		if authorizerId := cfg.GetString("api.authorizerId"); authorizerId != "" {
			invokeArn := awsutils.GetAuthorizerArn(restApiId, authorizerId).String()
			logrus.Infof("updating authorizer %s", authorizerId)
			if _, err := gatewaySvc.UpdateAuthorizer(&apigateway.UpdateAuthorizerInput{
				RestApiId:    &restApiId,
				AuthorizerId: &authorizerId,
				PatchOperations: []*apigateway.PatchOperation{
					{
						Op:    aws.String(apigateway.OpReplace),
						Path:  aws.String("/authorizerUri"),
						Value: &gatewayLambdaInvocationArn,
					},
				},
			}); err != nil {
				logrus.Error(err)
			}
			if err := addInvokePermission(function.AliasArn, invokeArn); err != nil {
				logrus.Error(err)
				continue
			}
		}

		logrus.Info("giving API Gateway invoke permissions")
	}

	return nil
}
