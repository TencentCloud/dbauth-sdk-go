// Package token provides functionality for token caching and management.
package token

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/sirupsen/logrus"
	"github.com/tencentcloud/dbauth-sdk-go/dbauth/internal/constants"
	"github.com/tencentcloud/dbauth-sdk-go/dbauth/internal/utils"
	"github.com/tencentcloud/dbauth-sdk-go/dbauth/model"
)

const MaxPasswordSize = 200

var logging = logrus.WithField("component", "token_cache")

// Cache represents a token cache.
type Cache struct {
	tokenMap sync.Map
}

// NewTokenCache creates a new token cache.
func NewTokenCache() *Cache {
	return &Cache{}
}

// GetAuthToken gets the authentication token from the cache.
func (tc *Cache) GetAuthToken(key string) *Token {
	if value, ok := tc.tokenMap.Load(key); ok {
		return value.(*Token)
	}
	return nil
}

// SetAuthToken sets the authentication token in the cache.
func (tc *Cache) SetAuthToken(key string, token *Token) {
	if key == "" || token == nil {
		return
	}
	tc.tokenMap.Store(key, token)
}

// RemoveAuthToken removes the authentication token from the cache.
func (tc *Cache) RemoveAuthToken(key string) {
	tc.tokenMap.Delete(key)
}

// Fallback gets the authentication token from the cache.
func (tc *Cache) Fallback(request *model.GenerateAuthenticationTokenRequest) *Token {
	inputFilePath := tc.generateInputFilePath(request)
	if inputFilePath == "" {
		return nil
	}

	// If the file exists, read the password from the file
	if fileInfo, err := os.Stat(inputFilePath); err == nil {
		logging.Infof("file name: %s, file size: %d", inputFilePath, fileInfo.Size())
		// If the file size is 0 or the file size is greater than 200, skip the file
		if fileInfo.Size() == 0 {
			return nil
		}

		if fileInfo.Size() > MaxPasswordSize {
			logging.Errorf("The file size is greater than 200, skip the file: %s", inputFilePath)
			return nil
		}

		lines, err := tc.readAllLines(inputFilePath)
		if err != nil {
			logging.Errorf("Failed to read all lines from the file: %s, error: %v", inputFilePath, err)
			return nil
		}

		if len(lines) != 1 {
			logging.Errorf("The file contains %d lines, expected exactly one line. Skipping file: %s",
				len(lines), inputFilePath)
			return nil
		}

		passwd := lines[0]

		logging.Infof("Reading the password from the file: %s", inputFilePath)
		return NewToken(passwd, utils.GetCurrentTimeMillis()+constants.MaxDelay)
	}

	return nil
}

func (tc *Cache) readAllLines(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return lines, nil
}

func (tc *Cache) generateInputFilePath(request *model.GenerateAuthenticationTokenRequest) string {
	region, instanceId, userName := request.Region(), request.InstanceId(), request.UserName()
	wd, err := os.Getwd()
	if err != nil {
		logging.Errorf("Failed to get working directory: %v", err)
		return ""
	}
	path := filepath.Join(constants.InputPathDir,
		region+constants.DELIMITER+instanceId+constants.DELIMITER+userName+".pwd")
	return filepath.Join(wd, path)
}
