语言 : [🇺🇸](./README.md) | 🇨🇳
<h1 align="center">Tencent Cloud DBAuth SDK</h1>
<div align="center">
欢迎使用腾讯云数据库CAM验证SDK，该SDK为开发者提供了支持的开发工具，以访问腾讯云数据库CAM验证服务，简化了腾讯云数据库CAM验证服务的接入过程。
</div>

### 依赖环境

1. 依赖环境：Go 1.17 版本及以上
2. 使用前需要在腾讯云控制台启用CAM验证。
3. 在腾讯云控制台[账号信息](https://console.cloud.tencent.com/developer)
   页面查看账号APPID，[访问管理](https://console.cloud.tencent.com/cam/capi)页面获取 SecretID 和 SecretKey 。

### 使用

```bash
go get -v -u github.com/tencentcloud/dbauth-sdk-go
```

#### 间接依赖项

tencentcloud-sdk-go v1.0.1015版本及以上。

### 示例 - 连接到数据库实例

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
	// 定义数据库连接参数
	region := "ap-guangzhou"
	instanceId := "cdb-123456"
	userName := "camtest"
	host := "gz-cdb-123456.sql.tencentcdb.com"
	port := 3306
	dbName := "test"
	ak := os.Getenv("TENCENTCLOUD_SECRET_ID")
	sk := os.Getenv("TENCENTCLOUD_SECRET_KEY")

	// 获取连接
	connection, err := getDBConnectionUsingCam(ak, sk, region, instanceId, userName, host, port, dbName)
	if err != nil {
		logrus.Error("Failed to get connection:", err)
		return
	}

	// 验证连接是否成功
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

	// 关闭连接
	if err := stmt.Close(); err != nil {
		logrus.Error("Failed to close statement:", err)
	}
	if err := connection.Close(); err != nil {
		logrus.Error("Failed to close connection:", err)
	}
}

// 使用CAM获取数据库连接
func getDBConnectionUsingCam(secretId, secretKey, region, instanceId, userName, host string, port int, dbName string) (*sql.DB, error) {
	credential := common.NewCredential(secretId, secretKey)
	maxAttempts := 3
	var lastErr error

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		// 获取认证Token
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

// 获取认证Token
func getAuthToken(region, instanceId, userName string, credential *common.Credential) (string, error) {
	// 实例化一个client选项，可选的，没有特殊需求可以跳过
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = "cam.tencentcloudapi.com"
	// 创建一个GenerateAuthenticationTokenRequest对象，ClientProfile是可选的
	tokenRequest, err := model.NewGenerateAuthenticationTokenRequest(region, instanceId, userName, credential, cpf)
	if err != nil {
		logrus.Errorf("Failed to create GenerateAuthenticationTokenRequest: %v", err)
		return "", err
	}

	return dbauth.GenerateAuthenticationToken(tokenRequest)
}

```

### 错误码

参见 [错误码](https://cloud.tencent.com/document/product/598/33168)。

### 局限性

使用 CAM 数据库身份验证时存在一些限制。以下内容来自 CAM
身份验证文档。

当您使用 CAM 数据库身份验证时，您的应用程序必须生成 CAM 身份验证令牌。然后，您的应用程序使用该令牌连接到数据库实例或集群。

我们建议如下：

* 使用 CAM 数据库身份验证作为临时、个人访问数据库的机制。
* 仅对可以轻松重试的工作负载使用 CAM 数据库身份验证。
