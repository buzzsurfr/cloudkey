package cmd

import (
	"fmt"

	"github.com/buzzsurfr/cloudkey/cloud/aws"
	"github.com/spf13/cobra"
)

// rotateCmd represents the rotate command
var rotateCmd = &cobra.Command{
	Use:   "rotate",
	Short: "Rotate the cloud access key",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("rotate called")
		p, err := aws.Current()
		if err != nil {
			panic(err)
		}
		fmt.Println("Profile: ", p.Name)
	},
}

func init() {
	rootCmd.AddCommand(rotateCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// rotateCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// rotateCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	rotateCmd.Flags().StringVarP(&profileName, "profile", "p", "default", "Profile to rotate")
}
