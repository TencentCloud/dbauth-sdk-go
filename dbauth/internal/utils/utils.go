// Package utils provides utility functions for the dbauth package.
package utils

import "time"

// GetCurrentTimeMillis returns the current time in milliseconds.
func GetCurrentTimeMillis() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}
