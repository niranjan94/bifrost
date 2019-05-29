package awsutils

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/spf13/viper"
	"sync"
)

var (
	awsSession *session.Session
	identity *sts.GetCallerIdentityOutput
	sessionOnce sync.Once
	identityOnce sync.Once
)

func GetSession() *session.Session {
	sessionOnce.Do(func() {
		awsSession = session.Must(session.NewSession(&aws.Config{
			Region: aws.String(viper.GetString("region")),
		}))
	})
	return awsSession
}

func GetIdentity() *sts.GetCallerIdentityOutput {
	identityOnce.Do(func() {
		svc := sts.New(GetSession())
		result, err := svc.GetCallerIdentity(&sts.GetCallerIdentityInput{})
		if err != nil {
			identity = nil
		}
		identity = result
	})
	return identity
}