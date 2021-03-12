// Usage: from OS console set environment variables
// e.g $ export AWS_REGION = "us-east-1; export PS_NAME="/apps/test/rds,myparameter,/apps/test/mytest,/apps/test/newtest,apptest"
// execute the command:
// $ go run main.go
// us-east-1
// [/apps/test/rds myparameter /apps/test/mytest /apps/test/newtest apptest]
// /apps/test/rds
// {
//   Parameter: {
//     ARN: "arn:aws:ssm:us-east-1:663409173557:parameter/apps/test/rds",
//     DataType: "text",
//     LastModifiedDate: 2020-09-05 06:17:18.868 +0000 UTC,
//     Name: "/apps/test/rds",
//     Type: "SecureString",
//     Value: "securetest",
//     Version: 1
//   }
// }
// /apps/test/rds
// stat /tmp/secrets: no such file or directory
// /tmp/secrets//apps/test
// /tmp/secrets//apps/test/rds
// myparameter
// {
//   Parameter: {
//     ARN: "arn:aws:ssm:us-east-1:663409173557:parameter/myparameter",
//     DataType: "text",
//     LastModifiedDate: 2020-09-05 06:13:27.46 +0000 UTC,
//     Name: "myparameter",
//     Type: "SecureString",
//     Value: "testvalue",
//     Version: 1
//   }
// }
// ................

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/aws/aws-sdk-go/service/ssm/ssmiface"
)

func main() {

	secretARN := os.Getenv("SECRET_ARN")

	value := strings.Split(secretARN, ",")

	for i := 0; i < len(value); i++ {

		var AWSRegion string
		var value_name string
		if arn.IsARN(value[i]) {
			arnobj, _ := arn.Parse(value[i])
			AWSRegion = arnobj.Region
			value_name = arnobj.Resource
		} else {
			fmt.Println("ARN Provided: ", value[i])
			fmt.Println("Not a valid ARN")
			os.Exit(1)
		}

		psName := strings.Split(value_name, "parameter")

		sess, _ := session.NewSession()
		svc := ssm.New(sess, &aws.Config{
			Region: aws.String(AWSRegion),
		})

		result := getPS(svc, psName[1])
		if len(result) != 0 {
			writeOutput(result, psName[1])
		}
	}
}

func getPS(svc ssmiface.SSMAPI, psName string) string {

	param := &ssm.GetParameterInput{
		Name:           aws.String(psName),
		WithDecryption: aws.Bool(true),
	}
	result, err := svc.GetParameter(param)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case ssm.ErrCodeAssociationDoesNotExist:
				fmt.Println(ssm.ErrCodeAssociationDoesNotExist, aerr.Error())
			case ssm.ErrCodeAlreadyExistsException:
				fmt.Print(ssm.ErrCodeAlreadyExistsException, aerr.Error())
			case ssm.ErrCodeInternalServerError:
				fmt.Println(ssm.ErrCodeInternalServerError, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			fmt.Println(err.Error())
		}
		return ""
	}
	return *result.Parameter.Value
}

func writeOutput(output string, psName string) {
	// Create Secrets Directory
	_, err := os.Stat("/tmp/secrets")
	if os.IsNotExist(err) {
		errDir := os.MkdirAll(psName, 0755)
		if errDir != nil {
			fmt.Println(err)
		}
	}
	subdir := filepath.Dir(psName)
	fullpath := "/tmp/secrets/" + subdir
	base := filepath.Base(psName)
	errDir1 := os.MkdirAll(fullpath, 0755)

	if errDir1 != nil {
		fmt.Println(err)
	}
	filename := fullpath + "/" + base
	// Insert the secret into the filename
	f, err := os.Create(filename)
	if err != nil {
		return
	}
	defer f.Close()

	f.WriteString(output)
}


