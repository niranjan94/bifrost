package aws

import (
	"github.com/niranjan94/bifrost/provision/aws/cognito"
	"github.com/niranjan94/bifrost/provision/aws/functions"
	"github.com/niranjan94/bifrost/provision/aws/gateway"
	"github.com/sirupsen/logrus"
)

func Provision()  {
	deploymentPackages := functions.Deploy(
		functions.Build(),
	)
	if err := gateway.IntegrateFunctions(deploymentPackages); err != nil {
		logrus.Error(err)
	}
	if err := gateway.DeployStage(); err != nil {
		logrus.Error(err)
	}
	if err := cognito.IntegrateFunctions(deploymentPackages); err != nil {
		logrus.Error(err)
	}
}