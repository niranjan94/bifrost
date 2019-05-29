package gateway

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/niranjan94/bifrost/config"
	awsutils "github.com/niranjan94/bifrost/utils/aws"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func DeployStage() error {
	restApiId := config.GetString("apiGateway.restApiId")
	if restApiId == "" {
		return nil
	}
	gatewaySvc := apigateway.New(awsutils.GetSession())
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
	return err
}