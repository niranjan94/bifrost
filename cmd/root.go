package cmd

import (
	"fmt"
	"github.com/mitchellh/go-homedir"
	"github.com/niranjan94/bifrost/config"
	"github.com/niranjan94/bifrost/utils"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"strings"
)

var (
	cfgFile string
	DryRun bool
	region string
	stage string
	functionOnly bool
	filter string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "bifrost",
	Short: "Your bridge to microservices on AWS",
	Long:  `Deploy microservices to AWS using lambda & API Gateway with easy using cobra.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initLogger, initConfig)

	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is ./bifrost.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&DryRun, "dry-run", "d", false, "dry run mode (default is false)")
	rootCmd.PersistentFlags().StringVarP(&stage, "stage", "s", "dev", "Stage to use (default is dev)")
	rootCmd.PersistentFlags().StringVarP(&region, "region", "r", "ap-southeast-1", "region (default is ap-southeast-1)")
	rootCmd.PersistentFlags().BoolVar(&functionOnly, "functions-only", false, "Deploy only functions")
	rootCmd.PersistentFlags().StringVar(&filter, "only", "", "Deploy only specific resources")

	utils.Must(viper.BindPFlags(rootCmd.PersistentFlags()))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		// Search config in home directory with name ".bifrost" (without extension).
		replacer := strings.NewReplacer(".", "_")
		viper.AutomaticEnv()
		viper.SetEnvPrefix("BIFROST")
		viper.SetEnvKeyReplacer(replacer)
		viper.AddConfigPath(home)
		viper.AddConfigPath(".")
		viper.SetConfigName(".bifrost")
		viper.SetConfigName("bifrost")
	}
	viper.AutomaticEnv() // read in environment variables that match
	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		logrus.Debug("Using config file: ", viper.ConfigFileUsed())
	} else {
		logrus.Error("No config file found.")
		os.Exit(1)
	}
	config.LoadDefaults()
}

// initLogger initializes the logrus instance
func initLogger() {
	logrus.SetFormatter(&logrus.TextFormatter{
		DisableLevelTruncation: true,
		DisableTimestamp:       true,
		QuoteEmptyFields:       true,
	})
	logrus.SetOutput(os.Stdout)
	logrus.SetLevel(logrus.DebugLevel)
}
