package errorcode

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
)

func TestIsUserNotificationRequired_AuthFailureError(t *testing.T) {
	err := &errors.TencentCloudSDKError{Code: "AuthFailure.InvalidSecretId"}
	assert.True(t, IsUserNotificationRequired(err))
}

func TestIsUserNotificationRequired_AuthFailureError_Lower(t *testing.T) {
	err := &errors.TencentCloudSDKError{Code: "authFailure.invalidSecretId"}
	assert.True(t, IsUserNotificationRequired(err))
}

func TestIsUserNotificationRequired_ResourceNotFoundError(t *testing.T) {
	err := &errors.TencentCloudSDKError{Code: "ResourceNotFound.DataFlowAuthClose"}
	assert.True(t, IsUserNotificationRequired(err))
}

func TestIsUserNotificationRequired_ResourceNotFoundError_Lower(t *testing.T) {
	err := &errors.TencentCloudSDKError{Code: "resourceNotFound.dataFlowAuthClose"}
	assert.True(t, IsUserNotificationRequired(err))
}

func TestIsUserNotificationRequired_NonAuthFailureError(t *testing.T) {
	err := &errors.TencentCloudSDKError{Code: "InvalidParameter"}
	assert.False(t, IsUserNotificationRequired(err))
}

func TestIsUserNotificationRequired_EmptyErrorCode(t *testing.T) {
	err := &errors.TencentCloudSDKError{Code: ""}
	assert.False(t, IsUserNotificationRequired(err))
}

func TestIsUserNotificationRequired_NilError(t *testing.T) {
	assert.False(t, IsUserNotificationRequired(nil))
}

func TestIsUserNotificationRequired_NonTencentCloudSDKError(t *testing.T) {
	err := &errors.TencentCloudSDKError{}
	assert.False(t, IsUserNotificationRequired(err))
}
