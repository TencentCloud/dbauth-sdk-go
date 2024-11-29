// Package parser provides functions to parse and decrypt authentication tokens.
package parser

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/tencentcloud/dbauth-sdk-go/dbauth/internal/constants"
	"github.com/tencentcloud/dbauth-sdk-go/dbauth/pb"
)

// ParseAuthToken parses the authentication token and returns the authentication token information.
func ParseAuthToken(instanceId, region, userName, token string) (*pb.AuthTokenInfo, error) {
	if instanceId == "" || region == "" || userName == "" || token == "" {
		return nil, errors.New("param empty")
	}

	// Generate encryption key
	seedKey := sha256Hash([]byte(instanceId + constants.DELIMITER + region + constants.DELIMITER + userName))

	key := seedKey[:32]
	iv := seedKey[33:49]

	// Decrypt AuthToken
	decToken, err := decrypt(token[64:], key, iv)
	if err != nil {
		return nil, err
	}

	// Compare if the token has been truncated
	tokenHash := sha256Hash(decToken)

	if token[:64] != tokenHash {
		return nil, errors.New("token not compare")
	}

	// Parse token
	return getAuthTokenInfo(decToken)
}

// getAuthTokenInfo parses the AuthTokenInfo from the decrypted token
func getAuthTokenInfo(decToken []byte) (*pb.AuthTokenInfo, error) {
	subToken := decToken[4:]

	var tokenInfo pb.AuthTokenInfo
	if err := proto.Unmarshal(subToken, &tokenInfo); err != nil {
		return nil, fmt.Errorf("failed to parse AuthTokenInfo: %w", err)
	}
	return &tokenInfo, nil
}

// sha256Hash calculates the SHA256 hash of the input
func sha256Hash(data []byte) string {
	return fmt.Sprintf("%x", sha256.Sum256(data))
}

// decrypt decrypts the input using the key and iv
func decrypt(encryptedData, key, iv string) ([]byte, error) {
	keyBytes := []byte(key)
	ivBytes := []byte(iv)

	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return nil, err
	}

	cipherText, err := base64Decode(encryptedData)
	if err != nil {
		return nil, err
	}

	if len(cipherText) < aes.BlockSize {
		return nil, errors.New("ciphertext too short")
	}

	mode := cipher.NewCBCDecrypter(block, ivBytes)
	decryptedPaddedPlaintext := make([]byte, len(cipherText))
	mode.CryptBlocks(decryptedPaddedPlaintext, cipherText)

	return unpad(decryptedPaddedPlaintext, aes.BlockSize)
}

// base64Decode decodes a base64 string
func base64Decode(data string) ([]byte, error) {
	data = strings.ReplaceAll(data, "-", "+")
	data = strings.ReplaceAll(data, "_", "/")

	// add padding characters if necessary
	mod4 := len(data) % 4
	if mod4 != 0 {
		data += strings.Repeat("=", 4-mod4)
	}

	return base64.StdEncoding.DecodeString(data)
}

func unpad(data []byte, blockSize int) ([]byte, error) {
	if len(data) == 0 {
		return nil, errors.New("invalid padding size")
	}
	padding := int(data[len(data)-1])
	if padding > blockSize || padding == 0 {
		return nil, errors.New("invalid padding size")
	}
	for i := 0; i < padding; i++ {
		if data[len(data)-1-i] != byte(padding) {
			return nil, errors.New("invalid padding")
		}
	}
	return data[:len(data)-padding], nil
}
