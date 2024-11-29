// Package dbauth provides functionalities for database authentication.
package dbauth

import (
	"github.com/sirupsen/logrus"
	"github.com/tencentcloud/dbauth-sdk-go/dbauth/internal/errorcode"
	"github.com/tencentcloud/dbauth-sdk-go/dbauth/internal/signer"
	"github.com/tencentcloud/dbauth-sdk-go/dbauth/internal/utils"
	"github.com/tencentcloud/dbauth-sdk-go/dbauth/model"
	errorcodes "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cam/v20190116"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
)

var logging = logrus.WithField("component", "dbauth")

// GenerateAuthenticationToken generates an authentication token based on the provided token request.
func GenerateAuthenticationToken(tokenRequest *model.GenerateAuthenticationTokenRequest) (string, error) {
	if tokenRequest == nil {
		return "", errors.NewTencentCloudSDKError(
			errorcodes.INVALIDPARAMETER_PARAMERROR, "The token request is invalid.", "")
	}

	// Create a new Signer with the provided token request.
	s := signer.New(*tokenRequest)
	// Get the authentication token from the cache.
	cachedToken := s.GetAuthTokenFromCache()
	if cachedToken != nil {
		if cachedToken.GetExpires() > utils.GetCurrentTimeMillis() {
			// If the token has not expired, return the token.
			return cachedToken.GetAuthToken(), nil
		}
	}

	err := s.BuildAuthToken()
	if err == nil {
		return s.GetAuthTokenFromCache().GetAuthToken(), nil
	} else {
		logging.Error("Error occurred while generating authentication token", err)
		if cachedToken != nil {
			if errorcode.IsUserNotificationRequired(err) {
				// If the error code requires user notification, return the error.
				return "", err
			}
			// If the error code does not require user notification, return the cached token.
			return cachedToken.GetAuthToken(), nil
		}
		// If there is no cached token, return the error.
		return "", err
	}
}
