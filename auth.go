package box

import (
	"context"
	"encoding/pem"
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"github.com/youmark/pkcs8"
)

// Auth module
type Auth struct {
	*Client
	Token *Token
}

type boxClaims struct {
	BoxSubType string `json:"box_sub_type"`
	jwt.StandardClaims
}

// Credentials represents the Authentication information
type Credentials struct {
	ClientID     string  `json:"clientID"`
	ClientSecret string  `json:"clientSecret"`
	AppAuth      AppAuth `json:"appAuth"`
	EnterpriseID string  `json:"enterpriseID"`
}

// AppAuth is used to authenticate an application
type AppAuth struct {
	PublicKeyID string `json:"publickeyID"`
	PrivateKey  string `json:"privateKey"`
	Passphrase  string `json:"passphrase"`
}

// IsAuthenticated tells if the client is authenticated
func (module *Auth) IsAuthenticated() bool {
	//TODO: Add expiration...
	return module.Token != nil && len(module.Token.AccessToken) > 0
}

// Authenticate authenticates with the given credentials
// Currently only AppAuth is supported
func (module *Auth) Authenticate(ctx context.Context, creds Credentials) (err error) {
	if module.IsAuthenticated() {
		return nil
	}

	jwtToken := jwt.NewWithClaims(jwt.GetSigningMethod("RS256"), boxClaims{
		"enterprise",
		jwt.StandardClaims {
			Audience: "https://api.box.com/oauth2/token",
			ExpiresAt: time.Now().Add(30*time.Second).Unix(),
			Id:        uuid.Must(uuid.NewRandom()).String(),
			Issuer:    creds.ClientID,
			Subject:   creds.EnterpriseID,
		},
	})
	jwtToken.Header["kid"] = creds.AppAuth.PublicKeyID

	pemBlock, _ := pem.Decode([]byte(creds.AppAuth.PrivateKey))
	rsaKey, err := pkcs8.ParsePKCS8PrivateKeyRSA(pemBlock.Bytes, []byte(creds.AppAuth.Passphrase))
	if err != nil {
		return fmt.Errorf("Failed to parse RSA Key: %s", err)
	}

	signedToken, err := jwtToken.SignedString(rsaKey)
	if err != nil {
		return fmt.Errorf("Failed to sign the token. Error: %s", err)
	}

	token := Token{}
	if _, err = module.Client.sendRequest(ctx, &requestOptions{
		Method:     "POST",
		Path:       "https://api.box.com/oauth2/token", 
		Parameters: map[string]string{
			"grant_type":    "urn:ietf:params:oauth:grant-type:jwt-bearer",
			"client_id":     creds.ClientID,
			"client_secret": creds.ClientSecret,
			"assertion":     signedToken,
		},
	}, &token); err != nil {
		return
	}
	// TODO: if we get an invalid_grant: https://github.com/box/box-node-sdk/blob/1d51f676b1323135891a70d470e2c9de97be9437/lib/token-manager.js#L216
	if token.TokenType == "bearer" {
		token.TokenType = "Bearer"
	}
	module.Token = &token
	return
}