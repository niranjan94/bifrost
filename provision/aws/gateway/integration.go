package gateway

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/aws/aws-sdk-go/service/apigatewayv2"
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
	if viper.GetBool("functions-only") {
		return nil
	}

	restApiId := config.GetString("apiGateway.restApiId")
	if restApiId == "" {
		return nil
	}

	wsApiId := config.GetString("apiGateway.wsApiId")

	logrus.Info("querying API Gateway")

	gatewaySvc := apigateway.New(awsutils.GetSession())
	wsGatewaySvc := apigatewayv2.New(awsutils.GetSession())

	_, err := gatewaySvc.GetRestApi(&apigateway.GetRestApiInput{
		RestApiId: &restApiId,
	})
	if err != nil {
		return err
	}

	if wsApiId != "" {
		_, err := wsGatewaySvc.GetApi(&apigatewayv2.GetApiInput{
			ApiId: &wsApiId,
		})
		if err != nil {
			return err
		}
	}

	var resources []*apigateway.Resource
	var wsResources []*apigatewayv2.Route

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

	if wsApiId != "" {
		if routes, err := wsGatewaySvc.GetRoutes(&apigatewayv2.GetRoutesInput{
			ApiId: &wsApiId,
		}); err != nil {
			return err
		} else {
			wsResources = routes.Items
		}
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

	getWsResourceByKey := func(routeKey string) *apigatewayv2.Route {
		for idx := range wsResources {
			resource := wsResources[idx]
			if *resource.RouteKey == routeKey {
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

		wsResources := cfg.GetStringSlice("api.wsResources")
		if singleWsResource := cfg.GetString("api.wsResource"); singleWsResource != "" {
			wsResources = append(wsResources, singleWsResource)
		}

		for _, resourceString := range resources {
			if resource := strings.Split(resourceString, ":"); len(resource) >= 2 {
				method := strings.ToUpper(resource[0])
				resourcePath := resource[1]
				invokeArn := awsutils.GetInvokeApiArn(restApiId, stage, method, path.Join(resourcePrefix, resourcePath)).String()
				gatewayResource := getResourceByPath(resourcePath)
				if gatewayResource != nil {
					logrus.Infof("updating integration for %s resource %s", method, resourcePath)
					if viper.GetBool("dryRun") {
						logrus.Warn("dry run mode. skipping update.")
						continue
					}

					if _, err := gatewaySvc.UpdateIntegration(&apigateway.UpdateIntegrationInput{
						RestApiId:  &restApiId,
						ResourceId: gatewayResource.Id,
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

		if wsApiId != "" {
			for _, resourceString := range wsResources {
				gatewayResource := getWsResourceByKey(resourceString)
				if gatewayResource != nil {
					logrus.Infof("updating integration for resource %s", gatewayResource.RouteKey)
					if viper.GetBool("dryRun") {
						logrus.Warn("dry run mode. skipping update.")
						continue
					}

					splitTarget := strings.Split(*gatewayResource.Target, "/")
					integrationId := splitTarget[1]

					newIntegration := &apigatewayv2.UpdateIntegrationInput{
						ApiId: &wsApiId,
						ConnectionType: aws.String("INTERNET"),
						ContentHandlingStrategy: aws.String("CONVERT_TO_TEXT"),
						IntegrationMethod: aws.String("POST"),
						IntegrationType: aws.String("AWS_PROXY"),
						IntegrationUri: &gatewayLambdaInvocationArn,
						IntegrationId: &integrationId,
						PassthroughBehavior: aws.String("WHEN_NO_MATCH"),
						TimeoutInMillis: aws.Int64(29000),
					}

					if existingIntegration, err := wsGatewaySvc.GetIntegration(&apigatewayv2.GetIntegrationInput{
						ApiId: &wsApiId,
						IntegrationId: &integrationId,
					}); err == nil {
						newIntegration.ConnectionId = existingIntegration.ConnectionId
						newIntegration.CredentialsArn = existingIntegration.CredentialsArn
						newIntegration.RequestParameters = existingIntegration.RequestParameters
						newIntegration.RequestTemplates = existingIntegration.RequestTemplates
						newIntegration.TemplateSelectionExpression = existingIntegration.TemplateSelectionExpression
						newIntegration.Description = existingIntegration.Description
					}
					if _, err := wsGatewaySvc.UpdateIntegration(newIntegration); err != nil {
						logrus.Error(err)
					}
					invokeArn := awsutils.GetInvokeWsApiArn(wsApiId, *gatewayResource.RouteKey).String()
					if err := addInvokePermission(function.AliasArn, invokeArn); err != nil {
						logrus.Error(err)
						continue
					}

				} else {
					logrus.Errorf("could not find API Gateway v2 resource %s for %s", resourceString, function.FunctionName)
				}
			}
		}

		if authorizerId := cfg.GetString("api.authorizerId"); authorizerId != "" {
			invokeArn := awsutils.GetAuthorizerArn(restApiId, authorizerId).String()
			logrus.Infof("updating authorizer %s", authorizerId)
			if viper.GetBool("dryRun") {
				logrus.Warn("dry run mode. skipping update.")
			} else {
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

		}

		if wsAuthorizerId := cfg.GetString("api.wsAuthorizerId"); wsApiId != "" && wsAuthorizerId != "" {
			invokeArn := awsutils.GetAuthorizerArn(wsApiId, wsAuthorizerId).String()
			logrus.Infof("updating authorizer %s", wsAuthorizerId)
			if viper.GetBool("dryRun") {
				logrus.Warn("dry run mode. skipping update.")
				continue
			}

			newAuthorizer := &apigatewayv2.UpdateAuthorizerInput{
				ApiId: &wsApiId,
				AuthorizerType: aws.String("REQUEST"),
			}

			existingAuthorizer, err := wsGatewaySvc.GetAuthorizer(&apigatewayv2.GetAuthorizerInput{
				ApiId: &wsApiId,
				AuthorizerId: &wsAuthorizerId,
			})
			if err != nil {
				logrus.Error(err)
				continue
			}

			newAuthorizer.Name = existingAuthorizer.Name
			newAuthorizer.AuthorizerResultTtlInSeconds = existingAuthorizer.AuthorizerResultTtlInSeconds
			newAuthorizer.AuthorizerCredentialsArn = existingAuthorizer.AuthorizerCredentialsArn
			newAuthorizer.AuthorizerType = existingAuthorizer.AuthorizerType
			newAuthorizer.IdentitySource = existingAuthorizer.IdentitySource
			newAuthorizer.IdentityValidationExpression = existingAuthorizer.IdentityValidationExpression
			newAuthorizer.AuthorizerId = existingAuthorizer.AuthorizerId
			newAuthorizer.AuthorizerUri = &gatewayLambdaInvocationArn

			if _, err := wsGatewaySvc.UpdateAuthorizer(newAuthorizer); err != nil {
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
