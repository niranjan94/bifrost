package functions

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/niranjan94/bifrost/utils/merge"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"io/ioutil"
	"strings"
)

type deploymentPackage struct {
	name        string
	deployName  string
	packageFile string
	config      *viper.Viper
}

func Deploy(deploymentPackages []*deploymentPackage) {

	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(viper.GetString("region")),
	}))
	lambdaSvc := lambda.New(sess)

	for _, deploymentPackage := range deploymentPackages {

		cfg := deploymentPackage.config
		stage := cfg.GetString("stage")

		shouldCreate := false

		logrus.Infof("deploying %s as %s", deploymentPackage.name, deploymentPackage.deployName)

		functionOutput, err := lambdaSvc.GetFunction(&lambda.GetFunctionInput{
			FunctionName: &deploymentPackage.deployName,
		})

		var existingFunction *lambda.FunctionConfiguration
		var existingFunctionTags = make(map[string]*string)

		functionInput := &lambda.CreateFunctionInput{
			FunctionName: &deploymentPackage.deployName,
			VpcConfig: &lambda.VpcConfig{},
			Environment: &lambda.Environment{
				Variables: map[string]*string{},
			},
			TracingConfig: &lambda.TracingConfig{},
			Layers: []*string{},
		}

		if err != nil {
			if awsErr, ok := err.(awserr.Error); !ok || awsErr.Code() != lambda.ErrCodeResourceNotFoundException {
				logrus.Error(awsErr)
				continue
			}
			existingFunction = &lambda.FunctionConfiguration{}
			existingFunctionTags = map[string]*string{}
			shouldCreate = true
		} else {
			existingFunction = functionOutput.Configuration
			existingFunctionTags = functionOutput.Tags
		}

		if existingFunction.FunctionName != nil {
			if err := merge.Merge(
				existingFunction,
				functionInput,
				merge.Bind{
					From: "Environment",
					To: "Environment",
					Translator: func(value interface{}) (i interface{}, e error) {
						environment := functionInput.Environment
						if value == nil {return environment, nil}
						if err := merge.Merge(value, environment); err != nil {
							return nil, err
						}
						return environment, nil
					},
				},
				merge.Bind{
					From: "VpcConfig",
					To: "VpcConfig",
					Translator: func(value interface{}) (i interface{}, e error) {
						vpcConfig := functionInput.VpcConfig
						if value == nil {return vpcConfig, nil}
						if err := merge.Merge(value, vpcConfig); err != nil {
							return nil, err
						}
						return vpcConfig, nil
					},
				},
				merge.Bind{
					From: "TracingConfig",
					To: "TracingConfig",
					Translator: func(value interface{}) (i interface{}, e error) {
						tracingConfig := functionInput.TracingConfig
						if value == nil {return tracingConfig, nil}
						if err := merge.Merge(value, tracingConfig); err != nil {
							return nil, err
						}
						return tracingConfig, nil
					},
				},
				merge.Bind{
					From: "Layers",
					To: "Layers",
					Translator: func(value interface{}) (i interface{}, e error) {
						layers := functionInput.Layers
						if inputLayers, ok := value.([]*lambda.Layer); ok {
							for idx := range inputLayers {
								layers = append(layers, inputLayers[idx].Arn)
							}
						}
						return layers, nil
					},
				},
			); err != nil {
				logrus.Error(err)
			}
		}

		securityGroupIds := cfg.GetStringSlice("vpcConfig.securityGroupIds")
		subnetIds := cfg.GetStringSlice("vpcConfig.subnetIds")

		if securityGroupIds != nil {
			functionInput.VpcConfig.SecurityGroupIds = aws.StringSlice(securityGroupIds)
		}

		if subnetIds != nil {
			functionInput.VpcConfig.SubnetIds = aws.StringSlice(subnetIds)
		}

		functionTags := cfg.GetStringMapString("tags")
		functionEnvironment := cfg.GetStringMapString("environment")
		functionRole := cfg.GetString("role")
		functionRuntime := cfg.GetString("runtime")
		functionMemorySize := cfg.GetInt64("memorySize")
		functionTimeout := cfg.GetInt64("timeout")
		functionHandler := cfg.GetString("handler")

		if functionTags != nil {
			for k, v := range functionTags {
				existingFunctionTags[strings.ToUpper(k)] = aws.String(v)
			}
		}

		functionInput.Tags = existingFunctionTags

		if functionEnvironment != nil {
			for k, v := range functionEnvironment {
				functionInput.Environment.Variables[strings.ToUpper(k)] = aws.String(v)
			}
		}

		if functionRole != "" {
			functionInput.Role = &functionRole
		}

		if functionMemorySize >= 128 {
			functionInput.MemorySize = &functionMemorySize
		}

		if functionRuntime != "" {
			functionInput.Runtime = &functionRuntime
		}

		if functionTimeout > 0 {
			functionInput.Timeout = &functionTimeout
		}

		if functionHandler != "" {
			functionInput.Handler = &functionHandler
		}

		functionArchiveContents, err := ioutil.ReadFile(deploymentPackage.packageFile)
		if err != nil {
			logrus.Error(err)
			continue
		}

		functionInput.Code = &lambda.FunctionCode{
			ZipFile: functionArchiveContents,
		}

		var deployed *lambda.FunctionConfiguration
		var deployedCodeSha256 *string

		if err := functionInput.Validate(); err != nil {
			logrus.Error(err)
			continue
		}

		if shouldCreate {
			deployed, err = lambdaSvc.CreateFunction(functionInput)
			if err != nil {
				logrus.Error(err)
				continue
			}
			deployedCodeSha256 = deployed.CodeSha256
		} else {
			deployed, err = lambdaSvc.UpdateFunctionCode(&lambda.UpdateFunctionCodeInput{
				ZipFile: functionInput.Code.ZipFile,
				FunctionName: functionInput.FunctionName,
				RevisionId: functionOutput.Configuration.RevisionId,
			})
			if err != nil {
				logrus.Error(err)
				continue
			}
			deployedCodeSha256 = deployed.CodeSha256
			configUpdate := &lambda.UpdateFunctionConfigurationInput{}
			if err = merge.Merge(functionInput, configUpdate); err != nil {
				logrus.Error(err)
				continue
			}
			configUpdate.RevisionId = deployed.RevisionId
			deployed, err = lambdaSvc.UpdateFunctionConfiguration(configUpdate)
			if err != nil {
				logrus.Error(err)
				continue
			}
			_, err = lambdaSvc.TagResource(&lambda.TagResourceInput{
				Resource: deployed.FunctionArn,
				Tags: functionInput.Tags,
			})
			if err != nil {
				logrus.Error(err)
				continue
			}
		}

		deployed, err = lambdaSvc.PublishVersion(&lambda.PublishVersionInput{
			CodeSha256: deployedCodeSha256,
			RevisionId: deployed.RevisionId,
			FunctionName: deployed.FunctionName,
			Description: &stage,
		})
		if err != nil {
			logrus.Error(err)
			continue
		}

		logrus.Infof("published version %s for %s", *deployed.Version, *deployed.FunctionName)

		alias, err := lambdaSvc.GetAlias(&lambda.GetAliasInput{
			FunctionName: deployed.FunctionName,
			Name: &stage,
		})
		if err != nil {
			if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() == lambda.ErrCodeResourceNotFoundException {
				alias, err = lambdaSvc.CreateAlias(&lambda.CreateAliasInput{
					FunctionName: deployed.FunctionName,
					FunctionVersion: deployed.Version,
					Name: &stage,
				})
				if err != nil {
					logrus.Error(err)
				}
			} else {
				logrus.Error(awsErr)
				continue
			}
		} else {
			alias, err = lambdaSvc.UpdateAlias(&lambda.UpdateAliasInput{
				FunctionName: deployed.FunctionName,
				FunctionVersion: deployed.Version,
				Name: &stage,
				RevisionId: alias.RevisionId,
			})
			if err != nil {
				logrus.Error(err)
			}
		}

		logrus.Infof("published alias %s for %s", *alias.Name, *deployed.FunctionName)
		logrus.Infof("deployed %s as %s", deploymentPackage.name, *deployed.FunctionName)
	}
}
