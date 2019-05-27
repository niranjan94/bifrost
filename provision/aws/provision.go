package aws

import "github.com/niranjan94/bifrost/provision/aws/functions"

func Provision()  {
	functions.Deploy(
		functions.Build(),
	)
}