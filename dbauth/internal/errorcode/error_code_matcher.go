// Package errorcode provides functions to handle error codes and determine if user notification is required.
package errorcode

import (
	"strings"

	cam "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cam/v20190116"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
)

const ErrorAuthFailurePrefix = "AuthFailure."

// IsUserNotificationRequired checks if the error code requires user notification.
func IsUserNotificationRequired(err error) bool {
	tcErr, ok := err.(*errors.TencentCloudSDKError)
	if !ok {
		return false
	}

	errorCode := tcErr.Code
	if errorCode == "" {
		return false
	}

	// ignoring case considerations
	lowerErrorCode := strings.ToLower(errorCode)
	return strings.HasPrefix(lowerErrorCode, strings.ToLower(ErrorAuthFailurePrefix)) ||
		strings.EqualFold(lowerErrorCode, cam.RESOURCENOTFOUND_DATAFLOWAUTHCLOSE)
}
