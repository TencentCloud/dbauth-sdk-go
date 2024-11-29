package main

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/tencentcloud/dbauth-sdk-go/dbauth"
	"github.com/tencentcloud/dbauth-sdk-go/dbauth/model"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
)

func init() {
	logrus.SetOutput(os.Stdout)
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	logrus.SetLevel(logrus.DebugLevel)
}

func main() {
	// Define parameters for Authentication Token
	var (
		region     = "ap-guangzhou"
		instanceId = "cdb-123456"
		userName   = "camtest"
		credential = common.NewCredential(os.Getenv("AK"), os.Getenv("SK"))
	)

	// Instantiate an HTTP profile, optional, can be skipped if there are no special requirements
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = "cam.tencentcloudapi.com"

	// Build a GenerateAuthenticationTokenRequest
	tokenRequest, err := model.NewGenerateAuthenticationTokenRequest(region, instanceId, userName, credential, cpf)
	if err != nil {
		logrus.Errorf("Failed to create GenerateAuthenticationTokenRequest: %v", err)
	}

	// Call the generateAuthenticationToken function
	authToken, err := dbauth.GenerateAuthenticationToken(tokenRequest)
	if err != nil {
		logrus.Error("Failed to generate authentication token: ", err)
	}

	logrus.Infof("Generated Authentication Token: %s", authToken)
}
