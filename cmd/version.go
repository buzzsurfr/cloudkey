/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	goVersion "go.hein.dev/go-version"
)

// versionCmd represents the version command
var (
	shortened   = false
	mainVersion = "dev"
	mainCommit  = "none"
	mainDate    = "unknown"
	output      = "json"
	versionCmd  = &cobra.Command{
		Use:   "version",
		Short: "Version will output the current build information",
		Long:  ``,
		Run: func(cmd *cobra.Command, args []string) {
			resp := goVersion.FuncWithOutput(shortened, mainVersion, mainCommit, mainDate, output)
			fmt.Print(resp)
			return
		},
	}
)

func init() {
	rootCmd.AddCommand(versionCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// versionCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// versionCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	versionCmd.Flags().BoolVarP(&shortened, "short", "s", false, "Print just the version number.")
	versionCmd.Flags().StringVarP(&output, "output", "o", "json", "Output format. One of 'yaml' or 'json'.")
}
