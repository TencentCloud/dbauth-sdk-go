// Package token provides structures and functions for managing authentication tokens.
package token

// Token represents an authentication token with its expiration time.
type Token struct {
	authToken string
	expires   int64
}

// NewToken creates a new Token with the provided authentication token and expiration time.
func NewToken(authToken string, expires int64) *Token {
	return &Token{authToken: authToken, expires: expires}
}

// GetAuthToken returns the authentication token.
func (t *Token) GetAuthToken() string {
	return t.authToken
}

// GetExpires returns the expiration time.
func (t *Token) GetExpires() int64 {
	return t.expires
}
