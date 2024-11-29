Language : 🇺🇸 | [🇨🇳](./README.zh-CN.md)
<h1 align="center">Tencent Cloud DBAuth SDK</h1>
<div align="center">
Welcome to the Tencent Cloud DBAuth SDK, which provides developers with supporting development tools to access the Tencent Cloud Database CAM verification service, simplifying the access process of the Tencent Cloud Database CAM verification service.
</div>

### Dependency Environment

1. Dependency Environment: Go 1.17 and above.
2. Before use, CAM verification must be enabled on the Tencent Cloud console.
3. On the Tencent Cloud console, view the account APPID on
   the [account information](https://console.cloud.tencent.com/developer) page, and obtain the SecretID and SecretKey on
   the [access management](https://console.cloud.tencent.com/cam/capi) page.

### USAGE

```bash
go get -v -u github.com/tencentcloud/dbauth-sdk-go
```

#### Indirect Dependencies

For tencentcloud-sdk-go v1.0.1015 and above.

### Example - Connect to a Database Instance

```
package main

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/sirupsen/logrus"
	"github.com/tencentcloud/dbauth-sdk-go/dbauth"
	"github.com/tencentcloud/dbauth-sdk-go/dbauth/model"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
)

func init() {
	logrus.SetOutput(os.Stdout)
	logrus.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})
	logrus.SetLevel(logrus.InfoLevel)
}

func main() {
	// Define parameters for Authentication Token
	region := "ap-guangzhou"
	instanceId := "cdb-123456"
	userName := "camtest"
	host := "gz-cdb-123456.sql.tencentcdb.com"
	port := 3306
	dbName := "test"
	ak := os.Getenv("TENCENTCLOUD_SECRET_ID")
	sk := os.Getenv("TENCENTCLOUD_SECRET_KEY")

	// Get the connection
	connection, err := getDBConnectionUsingCam(ak, sk, region, instanceId, userName, host, port, dbName)
	if err != nil {
		logrus.Error("Failed to get connection:", err)
		return
	}

	// Verify the connection is successful
	stmt, err := connection.Query("SELECT 'Success!';")
	if err != nil {
		logrus.Error("Failed to execute query:", err)
		return
	}
	for stmt.Next() {
		var result string
		stmt.Scan(&result)
		logrus.Info(result) // Success!
	}

	// Close the connection
	if err := stmt.Close(); err != nil {
		logrus.Error("Failed to close statement:", err)
	}
	if err := connection.Close(); err != nil {
		logrus.Error("Failed to close connection:", err)
	}
}

// Get a database connection using CAM Database Authentication
func getDBConnectionUsingCam(secretId, secretKey, region, instanceId, userName, host string, port int, dbName string) (*sql.DB, error) {
	credential := common.NewCredential(secretId, secretKey)
	maxAttempts := 3
	var lastErr error

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		// Get the authentication token using the credentials
		authToken, err := getAuthToken(region, instanceId, userName, credential)
		if err != nil {
			return nil, err
		}

		connectionUrl := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", userName, authToken, host, port, dbName)
		db, err := sql.Open("mysql", connectionUrl)
		if err != nil {
			lastErr = err
			logrus.Warnf("Open connection failed. Attempt %d failed.", attempt)
			time.Sleep(5 * time.Second)
			continue
		}
		if err = db.Ping(); err != nil {
			lastErr = err
			logrus.Warnf("Ping failed. Attempt %d failed.", attempt)
			time.Sleep(5 * time.Second)
			continue
		}
		return db, nil
	}

	logrus.Error("All attempts failed. error:", lastErr)
	return nil, lastErr
}

// Get an authentication token
func getAuthToken(region, instanceId, userName string, credential *common.Credential) (string, error) {
	// Instantiate a client profile, optional, can be skipped if there are no special requirements
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = "cam.tencentcloudapi.com"
	// Create a GenerateAuthenticationTokenRequest object, ClientProfile is optional
	tokenRequest, err := model.NewGenerateAuthenticationTokenRequest(region, instanceId, userName, credential, cpf)
	if err != nil {
		logrus.Errorf("Failed to create GenerateAuthenticationTokenRequest: %v", err)
		return "", err
	}

	return dbauth.GenerateAuthenticationToken(tokenRequest)
}

```

### Error Codes

Refer to the [error code document](https://cloud.tencent.com/document/product/598/33168) for more information.

### Limitations

There are some limitations when you use CAM database authentication. The following is from the CAM authentication
documentation.

When you use CAM database authentication, your application must generate an CAM authentication token. Your application
then uses that token to connect to the DB instance or cluster.

We recommend the following:

* Use CAM database authentication as a mechanism for temporary, personal access to databases.
* Use CAM database authentication only for workloads that can be easily retried.
