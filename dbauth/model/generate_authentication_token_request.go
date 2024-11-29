// Package model contains the data structures for the dbauth package.
package model

import (
	errorcodes "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cam/v20190116"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
)

// GenerateAuthenticationTokenRequest represents the request to generate an authentication token.
type GenerateAuthenticationTokenRequest struct {
	region        string
	instanceId    string
	userName      string
	credential    *common.Credential
	clientProfile *profile.ClientProfile
}

// NewGenerateAuthenticationTokenRequest creates a new GenerateAuthenticationTokenRequest.
func NewGenerateAuthenticationTokenRequest(region, instanceId, userName string,
	credential *common.Credential, clientProfile *profile.ClientProfile) (*GenerateAuthenticationTokenRequest, error) {

	if region == "" {
		return nil, errors.NewTencentCloudSDKError(
			errorcodes.INVALIDPARAMETER_RESOURCEREGIONERROR, "The region is invalid.", "")
	}
	if instanceId == "" {
		return nil, errors.NewTencentCloudSDKError(
			errorcodes.INVALIDPARAMETER_RESOURCEERROR, "The instanceId is invalid.", "")
	}
	if userName == "" {
		return nil, errors.NewTencentCloudSDKError(
			errorcodes.INVALIDPARAMETER_USERNAMEILLEGAL, "The userName is invalid.", "")
	}
	if credential == nil || credential.SecretId == "" || credential.SecretKey == "" {
		return nil, errors.NewTencentCloudSDKError(
			errorcodes.RESOURCENOTFOUND_SECRETNOTEXIST, "The credential is invalid.", "")
	}

	return &GenerateAuthenticationTokenRequest{
		region:        region,
		instanceId:    instanceId,
		userName:      userName,
		credential:    credential,
		clientProfile: clientProfile,
	}, nil
}

// Region returns the region.
func (r *GenerateAuthenticationTokenRequest) Region() string {
	return r.region
}

// InstanceId returns the instanceId.
func (r *GenerateAuthenticationTokenRequest) InstanceId() string {
	return r.instanceId
}

// UserName returns the userName.
func (r *GenerateAuthenticationTokenRequest) UserName() string {
	return r.userName
}

// Credential returns the credential.
func (r *GenerateAuthenticationTokenRequest) Credential() *common.Credential {
	return r.credential
}

// ClientProfile returns the clientProfile.
func (r *GenerateAuthenticationTokenRequest) ClientProfile() *profile.ClientProfile {
	return r.clientProfile
}
