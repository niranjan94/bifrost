package gateway

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/aws/aws-sdk-go/service/apigatewayv2"
	"github.com/niranjan94/bifrost/config"
	awsutils "github.com/niranjan94/bifrost/utils/aws"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func DeployStage() error {
	if viper.GetBool("functions-only") {
		return nil
	}

	restApiId := config.GetString("apiGateway.restApiId")
	if restApiId == "" {
		return nil
	}
	wsApiId := config.GetString("apiGateway.wsApiId")
	gatewaySvc := apigateway.New(awsutils.GetSession())
	wsGatewaySvc := apigatewayv2.New(awsutils.GetSession())

	stage := viper.GetString("defaults.stage")

	logrus.Info("deploying stage ", stage)

	if viper.GetBool("dryRun") {
		logrus.Warn("dry run mode. skipping deploy.")
		return nil
	}

	_, err := gatewaySvc.CreateDeployment(&apigateway.CreateDeploymentInput{
		RestApiId: &restApiId,
		StageName: aws.String(viper.GetString("defaults.stage")),
	})

	if wsApiId != "" {
		if _, err := wsGatewaySvc.CreateDeployment(&apigatewayv2.CreateDeploymentInput{
			ApiId: &wsApiId,
			StageName: aws.String(viper.GetString("defaults.stage")),
		}); err != nil {
			logrus.Error(err)
		}
	}
	return err
}