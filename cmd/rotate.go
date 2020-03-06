package cmd

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/sts"
	cloudAWS "github.com/buzzsurfr/cloudkey/cloud/aws"
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
	Run: rotateFunc,
}

func rotateFunc(cmd *cobra.Command, args []string) {
	// fmt.Println("rotate called")
	var p cloudAWS.Profile
	var err error
	if profileName != "" {
		p, err = cloudAWS.GetByName(profileName)
	} else {
		p, err = cloudAWS.Current()
	}
	if err != nil {
		panic(err)
	}
	// fmt.Printf("Profile: %+v\n", p)

	oldSess := p.Session()
	oldCred := p.Cred

	// List Access Keys
	userName, err := SessionUserName(oldSess)
	if err != nil {
		panic(err)
	}

	// Get Access Keys
	oldIamSvc := iam.New(oldSess)
	result, err := oldIamSvc.ListAccessKeys(&iam.ListAccessKeysInput{
		UserName: aws.String(userName),
	})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case iam.ErrCodeNoSuchEntityException:
				fmt.Println(iam.ErrCodeNoSuchEntityException, aerr.Error())
			case iam.ErrCodeServiceFailureException:
				fmt.Println(iam.ErrCodeServiceFailureException, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
		}
		return
	}
	// fmt.Printf("ListAccessKeys: %+v\n", result)
	if len(result.AccessKeyMetadata) != 1 {
		fmt.Println("Too many access keys")
		return
	}

	// Create new access key
	newAccessKey, err := oldIamSvc.CreateAccessKey(&iam.CreateAccessKeyInput{
		UserName: aws.String(userName),
	})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case iam.ErrCodeNoSuchEntityException:
				fmt.Println(iam.ErrCodeNoSuchEntityException, aerr.Error())
			case iam.ErrCodeLimitExceededException:
				fmt.Println(iam.ErrCodeLimitExceededException, aerr.Error())
			case iam.ErrCodeServiceFailureException:
				fmt.Println(iam.ErrCodeServiceFailureException, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
		}
		return
	}
	// fmt.Printf("CreateAccessKey: %+v\n", newAccessKey)

	// Create new credential from access key
	cred, err := cloudAWS.FromAccessKey(*newAccessKey.AccessKey)
	if err != nil {
		panic(err)
	}

	// Save cred to profile
	p.UpdateCredential(cred)

	// Create new AWS session
	newSess := p.Session()

	// Sleep for 15 seconds to allow access key to activate
	time.Sleep(15 * time.Second)

	// Deactivate old access key using new access key
	newIamSvc := iam.New(newSess)
	_, err = newIamSvc.UpdateAccessKey(&iam.UpdateAccessKeyInput{
		AccessKeyId: aws.String(oldCred.AccessKeyID),
		Status:      aws.String("Inactive"),
		UserName:    aws.String(userName),
	})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case iam.ErrCodeNoSuchEntityException:
				fmt.Println(iam.ErrCodeNoSuchEntityException, aerr.Error())
			case iam.ErrCodeLimitExceededException:
				fmt.Println(iam.ErrCodeLimitExceededException, aerr.Error())
			case iam.ErrCodeServiceFailureException:
				fmt.Println(iam.ErrCodeServiceFailureException, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
		}
		return
	}

	// Delete old access key using new access key
	_, err = newIamSvc.DeleteAccessKey(&iam.DeleteAccessKeyInput{
		AccessKeyId: aws.String(oldCred.AccessKeyID),
		UserName:    aws.String(userName),
	})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case iam.ErrCodeNoSuchEntityException:
				fmt.Println(iam.ErrCodeNoSuchEntityException, aerr.Error())
			case iam.ErrCodeLimitExceededException:
				fmt.Println(iam.ErrCodeLimitExceededException, aerr.Error())
			case iam.ErrCodeServiceFailureException:
				fmt.Println(iam.ErrCodeServiceFailureException, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
		}
		return
	}
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
	rotateCmd.Flags().StringVarP(&profileName, "profile", "p", "", "Profile to rotate")
}

// SessionUserName gets the user name of the current session
func SessionUserName(sess *session.Session) (string, error) {
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
