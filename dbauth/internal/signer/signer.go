// Package signer provides structures and functions for generating authentication tokens.
package signer

import (
	"encoding/base64"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/tencentcloud/dbauth-sdk-go/dbauth/internal/constants"
	"github.com/tencentcloud/dbauth-sdk-go/dbauth/internal/errorcode"
	"github.com/tencentcloud/dbauth-sdk-go/dbauth/internal/parser"
	"github.com/tencentcloud/dbauth-sdk-go/dbauth/internal/timer"
	"github.com/tencentcloud/dbauth-sdk-go/dbauth/internal/token"
	"github.com/tencentcloud/dbauth-sdk-go/dbauth/internal/utils"
	"github.com/tencentcloud/dbauth-sdk-go/dbauth/model"
	cam "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cam/v20190116"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
)

const tokenUpdateInterval = 5000

var (
	tokenCache   = token.NewTokenCache()
	timerManager = timer.NewManager()
	logging      = logrus.WithField("component", "signer")
)

// Signer represents the authentication token generation logic.
type Signer struct {
	authKey string
	request model.GenerateAuthenticationTokenRequest
}

// New creates a new Signer with the provided token request.
func New(request model.GenerateAuthenticationTokenRequest) *Signer {
	key := request.Region() + constants.DELIMITER + request.InstanceId() + constants.DELIMITER +
		request.UserName() + constants.DELIMITER + request.Credential().GetSecretId()
	authKey := base64.StdEncoding.EncodeToString([]byte(key))
	return &Signer{authKey: authKey, request: request}
}

// GetAuthTokenFromCache gets the authentication token from the cache.
func (s *Signer) GetAuthTokenFromCache() *token.Token {
	return tokenCache.GetAuthToken(s.authKey)
}

// BuildAuthToken generates the authentication token.
func (s *Signer) BuildAuthToken() error {
	logging.Debugf("Building authentication token for key")

	// 1. Request the authentication token
	authToken, err := s.getAuthToken()
	if err == nil {
		logging.Debugf("Successfully get the authentication token, expiry: %s",
			time.Unix(authToken.GetExpires()/1000, 0).Format("2006-01-02 15:04:05"))

		s.setTokenAndUpdateTask(authToken)
		return nil
	}

	// 2. If the error code requires user notification, return the error
	if errorcode.IsUserNotificationRequired(err) {
		return err
	}

	// 3. If the token generation fails, use the fallback token
	fallbackToken := tokenCache.Fallback(&s.request)
	if fallbackToken != nil {
		logging.Infof("Using the fallback token")
		s.setTokenAndUpdateTask(fallbackToken)
		return nil
	} else {
		// 4. If there is no fallback token, return the error
		return err
	}
}

func (s *Signer) setTokenAndUpdateTask(token *token.Token) {
	tokenCache.SetAuthToken(s.authKey, token)
	s.updateAuthTokenTask(token.GetExpires())
}

func (s *Signer) getAuthToken() (*token.Token, error) {
	response, err := s.requestAuthToken()
	if err != nil {
		return nil, err
	}

	if response == nil || response.Response == nil {
		return nil, s.logAndReturnError("Failed to request AuthToken, response is null",
			cam.INTERNALERROR, "")
	}

	requestId := *response.Response.RequestId
	tokenResponse := response.Response.Credentials
	if tokenResponse == nil || tokenResponse.Token == nil {
		return nil, s.logAndReturnError(
			fmt.Sprintf("Failed to request AuthToken, requestId: %v, tokenResponse is null", requestId),
			cam.INTERNALERROR, requestId)
	}

	// Decrypt the authToken
	authToken, err := s.decryptAuthToken(*tokenResponse.Token)
	if err != nil || authToken == "" {
		return nil, s.logAndReturnError(fmt.Sprintf("Failed to decrypt AuthToken, requestId: %v, error: %v",
			requestId, err), cam.INTERNALERROR, requestId)
	}

	// Calculate the expiry time of the authToken
	expiry := s.calculateExpiry(*tokenResponse.CurrentTime, *tokenResponse.NextRotationTime)
	return token.NewToken(authToken, expiry), nil
}

func (s *Signer) logAndReturnError(message, code, requestId string) error {
	logging.Errorf(message)
	return errors.NewTencentCloudSDKError(code, message, requestId)
}

func (s *Signer) decryptAuthToken(encAuthToken string) (string, error) {
	instanceId, region, userName := s.request.InstanceId(), s.request.Region(), s.request.UserName()
	tokenInfo, err := parser.ParseAuthToken(instanceId, region, userName, encAuthToken)
	if err != nil {
		return "", err
	}
	return tokenInfo.Password, nil
}

func (s *Signer) calculateExpiry(camServerTime, authTokenExpires int64) int64 {
	if authTokenExpires < camServerTime {
		return utils.GetCurrentTimeMillis() + tokenUpdateInterval
	}
	return utils.GetCurrentTimeMillis() + (authTokenExpires - camServerTime)
}

func (s *Signer) requestAuthToken() (*cam.BuildDataFlowAuthTokenResponse, error) {
	clientProfile := s.request.ClientProfile()
	if clientProfile == nil {
		clientProfile = profile.NewClientProfile()
		clientProfile.HttpProfile.Endpoint = constants.CamEndPoint
		clientProfile.HttpProfile.ReqTimeout = 30 // Set the request timeout to 30 seconds
	}

	client, err := cam.NewClient(s.request.Credential(), s.request.Region(), clientProfile)
	if err != nil {
		return nil, errors.NewTencentCloudSDKError(cam.INTERNALERROR,
			fmt.Sprintf("Failed to create the client, error: %v", err), "")
	}

	resourceId, region, userName := s.request.InstanceId(), s.request.Region(), s.request.UserName()
	req := cam.NewBuildDataFlowAuthTokenRequest()
	req.ResourceId = &resourceId
	req.ResourceRegion = &region
	req.ResourceAccount = &userName

	var lastErr error
	for i := 0; i < 3; i++ {
		resp, err := client.BuildDataFlowAuthToken(req)
		if err == nil {
			return resp, nil
		}

		if tcErr, ok := err.(*errors.TencentCloudSDKError); ok {
			lastErr = tcErr
			if errorcode.IsUserNotificationRequired(err) {
				logging.Errorf("Failed to request AuthToken, error: %s", tcErr.Message)
				break
			}
			logging.Errorf("Failed to request AuthToken, Retry, TencentCloudSDKError: %s", tcErr.Message)
		} else {
			logging.Errorf("Failed to request AuthToken, Retry, error: %v", err)
			lastErr = errors.NewTencentCloudSDKError(cam.INTERNALERROR,
				fmt.Sprintf("Failed to request AuthToken, error: %v", err), "")
		}
	}
	return nil, lastErr
}

func (s *Signer) updateAuthTokenTask(authTokenExpiry int64) {
	// Calculate the remaining time before the token expires
	remainingTimeBeforeExpiry := authTokenExpiry - utils.GetCurrentTimeMillis()
	// Get the delay for the next token update
	delayForNextTokenUpdate := remainingTimeBeforeExpiry
	if delayForNextTokenUpdate > tokenUpdateInterval {
		delayForNextTokenUpdate = tokenUpdateInterval
	}

	logging.Debugf("Scheduling next token key update in %v ms", delayForNextTokenUpdate)

	// Save the timer for the next token update
	timerManager.SaveTimer(s.authKey, delayForNextTokenUpdate, func() {
		err := s.BuildAuthToken()
		if err != nil {
			if errorcode.IsUserNotificationRequired(err) {
				// If a user notification is required, remove the token from the cache
				logging.Errorf("Failed to update the authentication token, error: %v", err)
				tokenCache.RemoveAuthToken(s.authKey)
				return
			}
			// If an internal error occurs, try to update the token again
			logging.Errorf("Failed to update the authentication token, Retry to update the token, error: %v", err)
			s.updateAuthTokenTask(utils.GetCurrentTimeMillis() + tokenUpdateInterval)
		}
	})
}
