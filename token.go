package box

import (
	"context"
	"encoding/json"
	"time"
)

// Token is the token used to send requests to Box.com
type Token struct {
	TokenType    string    `json:"token_type"`
	AccessToken  string    `json:"access_token"`
	ExpiresOn    time.Time `json:"expires_on"`
	RestrictedTo []string  `json:"restricted_to"`
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

// IsValid tells if the token is valid and not expired
func (token Token) IsValid() bool {
	return len(token.AccessToken) > 0 && time.Now().UTC().Before(token.ExpiresOn)
}

// UnmarshalJSON decodes JSON
func (token *Token) UnmarshalJSON(payload []byte) (err error) {
	type surrogate Token
	var inner struct {
		surrogate
		ExpiresIn int64 `json:"expires_in"`
	}
	if err = json.Unmarshal(payload, &inner); err != nil {
		return err
	}
	*token = Token(inner.surrogate)
	if token.TokenType == "bearer" {
		token.TokenType = "Bearer"
	}
	if inner.ExpiresIn > 0 {
		token.ExpiresOn = time.Now().UTC().Add(time.Duration(inner.ExpiresIn) * time.Second)
	}
	return
}