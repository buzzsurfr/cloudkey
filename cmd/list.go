package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/buzzsurfr/cloudkey/cloud/aws"
	"github.com/mattn/go-colorable"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Lists all cloud access keys",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: listFunc,
}

func listFunc(cmd *cobra.Command, args []string) {
	// fmt.Println("list called")
	var profiles aws.Profiles

	// Check for and add environment variable credentials
	envProfile, err := aws.FromEnviron()
	if err == nil { // we found a profile in env
		profiles.Profiles = append(profiles.Profiles, envProfile)
	}

	// Parse ~/.aws/credentials file (INI format) for profiles and credentials
	configProfiles, err := aws.FromConfigFile(err != nil)
	if err == nil { // we found profile(s) in config file
		for _, p := range configProfiles.Profiles {
			profiles.Profiles = append(profiles.Profiles, p)
		}
	}

	renderTable(profiles.Profiles) // DEBUG

	// Get Current username if none provided
	// sess := session.New()
	// currentUserName, err := getSessionContext(sess)
	// svc := iam.New(sess)
	// result, err := svc.ListAccessKeys(&iam.ListAccessKeysInput{
	// 	UserName: aws.String(currentUserName),
	// })
	// if err != nil {
	// 	if aerr, ok := err.(awserr.Error); ok {
	// 		switch aerr.Code() {
	// 		case iam.ErrCodeNoSuchEntityException:
	// 			fmt.Println(iam.ErrCodeNoSuchEntityException, aerr.Error())
	// 		case iam.ErrCodeServiceFailureException:
	// 			fmt.Println(iam.ErrCodeServiceFailureException, aerr.Error())
	// 		default:
	// 			fmt.Println(aerr.Error())
	// 		}
	// 	} else {
	// 		// Print the error, cast err to awserr.Error to get the Code and
	// 		// Message from an error.
	// 		fmt.Println(err.Error())
	// 	}
	// 	return
	// }
	// fmt.Printf("%s\n", result)
}

func getSessionContext(sess *session.Session) (string, error) {
	var userName string

	// AWS sts:GetCallerIdentity API
	svc := sts.New(sess)
	result, err := svc.GetCallerIdentity(&sts.GetCallerIdentityInput{})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
		}
		return "", err
	}

	// Parse ARN
	resultArn, err := arn.Parse(*result.Arn)
	if err != nil {
		return "", err
	}

	// Verify is a user
	s := strings.Split(resultArn.Resource, "/")
	if s[0] != "user" {
		return "", errors.New("Not a user")
	}
	userName = s[1]

	return userName, nil
}

func renderTable(profiles []aws.Profile) error {
	table := tablewriter.NewWriter(colorable.NewColorableStdout())
	table.SetHeader([]string{"Cloud", "Name", "Access Key ID", "Source"})
	table.SetAutoWrapText(false)
	table.SetAutoFormatHeaders(true)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("")
	table.SetHeaderLine(false)
	table.SetBorder(false)
	table.SetTablePadding("   ") // pad with tabs
	table.SetNoWhiteSpace(true)

	for _, profile := range profiles {
		if profile.IsCurrent {
			table.Rich([]string{
				profile.Cloud,
				profile.Name,
				obfuscateString(profile.Cred.AccessKeyID, 4),
				profile.Source,
			}, []tablewriter.Colors{
				tablewriter.Color(tablewriter.FgYellowColor),
				tablewriter.Color(tablewriter.FgYellowColor),
				tablewriter.Color(tablewriter.FgYellowColor),
				tablewriter.Color(tablewriter.FgYellowColor),
			})
		} else {
			table.Append([]string{
				profile.Cloud,
				profile.Name,
				obfuscateString(profile.Cred.AccessKeyID, 4),
				profile.Source,
			})
		}
	}
	table.Render()

	return nil
}

func obfuscateString(s string, n int) string {
	var ret string
	for i, v := range s {
		if i >= n && i < len(s)-n {
			ret += "*"
		} else {
			ret += string(v)
		}
	}
	return ret
}

func init() {
	rootCmd.AddCommand(listCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// listCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// listCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
