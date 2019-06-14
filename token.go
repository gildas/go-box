package box

import (
	"context"
)

// Token is the token used to send requests to Box.com
type Token struct {
	TokenType    string   `json:"token_type"`
	AccessToken  string   `json:"access_token"`
	ExpiresIn    int64    `json:"expires_in"` // TODO: We should transcript this in ExpiresAt time.Time
	RestrictedTo []string `json:"restricted_to"`
}

type key int

// TokenContextKey is the key for the token stored in a context.Context
const TokenContextKey key = iota

// ToContext stores this token to the given context
// If the token is not set, the context is untouched
func (token *Token) ToContext(parent context.Context) context.Context {
	if token == nil || len(token.AccessToken) == 0 {
		return parent
	}
	return context.WithValue(parent, TokenContextKey, token)
}

// TokenFromContext retrieves a token from the given context
// If no token was stored in the context, nil is returned
func TokenFromContext(ctx context.Context) (*Token) {
	value := ctx.Value(TokenContextKey)
	if value == nil {
		return nil
	}
	return value.(*Token)
}