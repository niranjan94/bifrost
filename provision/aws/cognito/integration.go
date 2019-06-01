package cognito

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/niranjan94/bifrost/config"
	"github.com/niranjan94/bifrost/provision/aws/functions"
	"github.com/niranjan94/bifrost/utils"
	awsutils "github.com/niranjan94/bifrost/utils/aws"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"reflect"
)

func addInvokePermission(aliasArn string, poolArn string) error {
	lambdaSvc := lambda.New(awsutils.GetSession())
	statementId := utils.SHA1Hash(poolArn)

	_, _ = lambdaSvc.RemovePermission(&lambda.RemovePermissionInput{
		FunctionName: &aliasArn,
		StatementId:  &statementId,
	})
	_, err := lambdaSvc.AddPermission(&lambda.AddPermissionInput{
		FunctionName: &aliasArn,
		Principal:    aws.String("cognito-idp.amazonaws.com"),
		Action:       aws.String("lambda:InvokeFunction"),
		StatementId:  &statementId,
		SourceArn:    &poolArn,
	})
	return err
}

func IntegrateFunctions(functions []*functions.DeploymentPackage) error {
	if viper.GetBool("functions-only") {
		return nil
	}

	userPools := config.GetStringMapString("cognito.userPools")
	logrus.Info(userPools)
	if userPools == nil || len(userPools) == 0 {
		return nil
	}

	cognitoSvc := cognitoidentityprovider.New(awsutils.GetSession())

	for _, function := range functions {

		cfg := function.Config
		triggers := cfg.GetStringSlice("cognito.triggers")
		stage := cfg.GetString("stage")

		if len(triggers) == 0 {
			continue
		}

		poolId, exists := userPools[stage]
		if !exists {
			continue
		}

		logrus.Infof("activating %s on cognito pool %s", function.FunctionName, poolId)

		userPool, err := cognitoSvc.DescribeUserPool(&cognitoidentityprovider.DescribeUserPoolInput{
			UserPoolId: &poolId,
		})
		if err != nil {
			logrus.Error(err)
			continue
		}

		lambdaConfig := userPool.UserPool.LambdaConfig
		lambdaConfigElem := reflect.ValueOf(lambdaConfig).Elem()

		for idx := range triggers {
			triggerField := lambdaConfigElem.FieldByName(triggers[idx])
			if !triggerField.IsValid() || !triggerField.CanSet() {
				logrus.Info(triggerField.IsValid(), triggerField.CanSet())
				continue
			}
			triggerField.Set(reflect.ValueOf(&function.AliasArn))
		}

		if viper.GetBool("dryRun") {
			logrus.Warn("dry run mode. skipping activate.")
			continue
		}

		if _, err := cognitoSvc.UpdateUserPool(&cognitoidentityprovider.UpdateUserPoolInput{
			UserPoolId: &poolId,
			LambdaConfig: lambdaConfig,
		}); err != nil {
			logrus.Error(err)
			continue
		}

		logrus.Info("giving cognito invoke permissions")

		if err := addInvokePermission(function.AliasArn, *userPool.UserPool.Arn); err != nil {
			logrus.Error(err)
		}
	}
	return nil
}