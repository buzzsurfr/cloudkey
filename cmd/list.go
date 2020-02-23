package cmd

import (
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
	awsTypes "github.com/buzzsurfr/cloudkey/cloud/aws/types"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/mitchellh/mapstructure"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
	Run: func(cmd *cobra.Command, args []string) {
		// fmt.Println("list called")
		var profiles awsTypes.Profiles

		// Check for and add environment variable credentials
		envProfile, err := getProfileFromEnv()
		if err == nil { // we found a profile in env
			profiles.Profiles = append(profiles.Profiles, envProfile)
		}

		// Parse ~/.aws/credentials file (TOML format) for profiles and credentials
		configProfiles, err := getProfilesFromConfig(err != nil)
		if err == nil { // we found profile(s) in config file
			for _, p := range configProfiles.Profiles {
				profiles.Profiles = append(profiles.Profiles, p)
			}
		}

		printTable(profiles.Profiles) // DEBUG
		// fmt.Printf("%+v\n", profiles) // DEBUG

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
	},
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

func getCurrentProfile() string {
	currentProfile := "default"

	// Check environment variables
	profileVars := []string{
		"AWS_DEFAULT_PROFILE",
		"AWS_PROFILE",
	}
	for _, env := range profileVars {
		v, ok := os.LookupEnv(env)
		if ok {
			currentProfile = v
		}

	}
	return currentProfile
}

func getProfilesFromConfig(findDefault bool) (awsTypes.Profiles, error) {
	var profiles awsTypes.Profiles

	awsConfigPath, err := getConfigPath()
	if err != nil {
		// Returning profiles since it's empty here
		return profiles, err
		// May change later since this assumes no credentials file found in ~/.aws
	}

	// Parse AWS config file
	v := viper.New()
	v.SetConfigName("credentials")
	v.SetConfigType("ini")
	v.AddConfigPath(awsConfigPath)
	err = v.ReadInConfig()
	if err != nil {
		return awsTypes.Profiles{}, err
	}
	// fmt.Printf("Configuration file:\n%+v\n\n", v) // DEBUG
	allSettings := v.AllSettings()

	var currentProfile string
	if findDefault {
		currentProfile = getCurrentProfile()
	}

	for key, value := range allSettings {
		// fmt.Printf("Key: %s\nValue: %+v\n\n", key, value) // DEBUG
		var cred awsTypes.Credential
		mapstructure.Decode(value, &cred)
		profiles.Profiles = append(profiles.Profiles, awsTypes.Profile{
			Name:      key,
			Cloud:     "aws",
			Cred:      cred,
			Source:    "ConfigFile",
			IsCurrent: findDefault && currentProfile == key,
		})
	}

	// Sort by profile name
	sort.Slice(profiles.Profiles, func(i, j int) bool { return profiles.Profiles[i].Name < profiles.Profiles[j].Name })

	return profiles, err
}

func getProfileFromEnv() (awsTypes.Profile, error) {
	if c, ok := getEnviron(); ok {
		return awsTypes.Profile{
			Name:      "",
			Cloud:     "aws",
			Cred:      c,
			Source:    "EnvironmentVariable",
			IsCurrent: true,
		}, nil
	}
	return awsTypes.Profile{}, errors.New("No credential found in environment variable")

}

func getConfigPath() (string, error) {
	hd, err := homedir.Dir()
	if err != nil {
		return "", err
	}
	hde, _ := homedir.Expand(hd)
	if err != nil {
		return "", err
	}
	configPath := hde + string(os.PathSeparator) + ".aws"
	// fmt.Printf("AWS Config File directory: %s\n", configPath) // DEBUG
	return configPath, nil
}

func getEnviron() (awsTypes.Credential, bool) {
	if _, snok := os.LookupEnv("AWS_SESSION_NAME"); snok {
		return awsTypes.Credential{}, false
	}
	akid, akok := os.LookupEnv("AWS_ACCESS_KEY_ID")
	sak, skok := os.LookupEnv("AWS_SECRET_ACCESS_KEY")
	if akok && skok {
		return awsTypes.Credential{
			AccessKeyID:     akid,
			SecretAccessKey: sak,
		}, true
	}

	return awsTypes.Credential{}, false
}

func printTable(profiles []awsTypes.Profile) error {
	table := tablewriter.NewWriter(os.Stdout)
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
