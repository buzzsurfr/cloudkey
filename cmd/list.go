package cmd

import (
	"fmt"
	"sort"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
	cloudAWS "github.com/buzzsurfr/cloudkey/cloud/aws"
	"github.com/mattn/go-colorable"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Lists all cloud access keys",
	Long: `List pulls the credentials from environment variables and the credentials file
and outputs them into a table. The "active" profile (which will be rotated by
default or used with AWS CLI commands) will be in yellow text.

Example Output:
CLOUD   NAME      ACCESS KEY ID          SOURCE
aws               AKIA************MPLE   EnvironmentVariable
aws     default   AKIA************G7UP   ConfigFile
`,
	Run: listFunc,
}

func listFunc(cmd *cobra.Command, args []string) {
	// fmt.Println("list called")
	var profiles cloudAWS.Profiles

	// Check for and add environment variable credentials
	envProfile, err := cloudAWS.FromEnviron()
	if err == nil { // we found a profile in env
		profiles.Profiles = append(profiles.Profiles, envProfile)
	}

	// Parse ~/.aws/credentials file (INI format) for profiles and credentials
	configProfiles, err := cloudAWS.FromConfigFile(err != nil)
	if err == nil { // we found profile(s) in config file
		pChan := make(chan struct{})
		for _, p := range configProfiles.Profiles {
			go func(profile cloudAWS.Profile) {
				if output == "wide" {
					err := profile.Lookup()
					if err != nil {
						// Warn that this profile couldn't be looked up, but continue
						fmt.Println(err)
					}
				}
				profiles.Profiles = append(profiles.Profiles, profile)
				pChan <- struct{}{}
			}(p)
		}
		for _ = range configProfiles.Profiles {
			<-pChan
		}
	}

	// Sort by profile name
	sort.Slice(profiles.Profiles, func(i, j int) bool { return profiles.Profiles[i].Name < profiles.Profiles[j].Name })

	renderTable(profiles.Profiles) // DEBUG
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
	userName, err = UserName(aws.StringValue(result.Arn))
	if err != nil {
		return "", err
	}
	return userName, nil
}

func renderTable(profiles []cloudAWS.Profile) error {
	table := tablewriter.NewWriter(colorable.NewColorableStdout())
	headers := make([]string, 0)
	switch output {
	case "wide":
		headers = []string{"Cloud", "Name", "Account", "UserName", "Access Key ID", "Source"}
	default:
		headers = []string{"Cloud", "Name", "Access Key ID", "Source"}
	}
	table.SetHeader(headers)
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
		switch output {
		case "wide":
			userName, _ := UserName(aws.StringValue(profile.Arn))
			// Eventually we'll log errors to debug, but this isn't a big deal
			// if err != nil {
			// 	fmt.Println(err)
			// }
			if profile.IsCurrent {
				table.Rich([]string{
					profile.Cloud,
					profile.Name,
					aws.StringValue(profile.Account),
					userName,
					obfuscateString(profile.Cred.AccessKeyID, 4),
					profile.Source,
				}, []tablewriter.Colors{
					tablewriter.Color(tablewriter.FgYellowColor),
					tablewriter.Color(tablewriter.FgYellowColor),
					tablewriter.Color(tablewriter.FgYellowColor),
					tablewriter.Color(tablewriter.FgYellowColor),
					tablewriter.Color(tablewriter.FgYellowColor),
					tablewriter.Color(tablewriter.FgYellowColor),
				})
			} else {
				table.Append([]string{
					profile.Cloud,
					profile.Name,
					aws.StringValue(profile.Account),
					userName,
					obfuscateString(profile.Cred.AccessKeyID, 4),
					profile.Source,
				})
			}
		default:
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
	listCmd.Flags().StringVarP(&output, "output", "o", "table", "Output format. Only supports 'wide'.")

}
