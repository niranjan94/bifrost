package cmd

import (
	"fmt"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"strings"
)

var cfgFile string
var DryRun bool

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "bifrost",
	Short: "Your bridge to microservices on AWS",
	Long: `Deploy microservices to AWS using lambda & API Gateway with easy using cobra.`,
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
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.bifrost.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&DryRun, "dry-run", "d", false, "dry run mode (default is false)")
	if err := viper.BindPFlag("dryRun", rootCmd.PersistentFlags().Lookup("dry-run")); err != nil {
		panic(err)
	}
	if err := viper.BindPFlag("config", rootCmd.PersistentFlags().Lookup("config")); err != nil {
		panic(err)
	}
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
	}
	viper.AutomaticEnv() // read in environment variables that match
	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
