package cmd

import (
	"github.com/niranjan94/bifrost/provision/aws/functions"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// deployCmd represents the deploy command
var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy your stack to the cloud",
	Long:  `Deploy your stack to the cloud`,
	Run: func(cmd *cobra.Command, args []string) {
		logrus.Info("deploy called")
		functions.Deploy(
			functions.Build(),
		)
	},
}

func init() {
	rootCmd.AddCommand(deployCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// deployCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// deployCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
