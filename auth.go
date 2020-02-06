package box

import (
	"context"
	"encoding/pem"
	"encoding/json"
	"net/url"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gildas/go-errors"
	"github.com/gildas/go-request"
	"github.com/google/uuid"
	"github.com/youmark/pkcs8"
)

// Auth module
type Auth struct {
	*Client
	api   *url.URL
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
	EnterpriseID string  `json:"enterpriseID,omitempty"`
}

// AppAuth is used to authenticate an application
type AppAuth struct {
	PublicKeyID string `json:"publickeyID"`
	PrivateKey  string `json:"privateKey"`
	Passphrase  string `json:"passphrase"`
}

// IsAuthenticated tells if the client is authenticated
func (module *Auth) IsAuthenticated() bool {
	return module.Token != nil && module.Token.IsValid()
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
	if pemBlock == nil || len(pemBlock.Bytes) == 0 {
		return errors.Unauthorized.Wrap(InvalidPrivateKey)
	}
	rsaKey, err := pkcs8.ParsePKCS8PrivateKeyRSA(pemBlock.Bytes, []byte(creds.AppAuth.Passphrase))
	if err != nil {
		return errors.Unauthorized.Wrap(errors.WithMessage(InvalidPrivateKey, err.Error()))
	}

	signedToken, err := jwtToken.SignedString(rsaKey)
	if err != nil {
		return errors.Unauthorized.Wrap(err)
	}

	token := Token{}
	_, err = module.Client.sendRequest(ctx, &request.Options{
		URL:     module.api,
		Payload: map[string]string{
			"grant_type":    "urn:ietf:params:oauth:grant-type:jwt-bearer",
			"client_id":     creds.ClientID,
			"client_secret": creds.ClientSecret,
			"assertion":     signedToken,
		},
	}, &token)
	// TODO: if we get an invalid_grant: https://github.com/box/box-node-sdk/blob/1d51f676b1323135891a70d470e2c9de97be9437/lib/token-manager.js#L216
	module.Token = &token
	return
}

// MarshalJSON marshals into JSON
func (creds *Credentials) MarshalJSON() ([]byte, error) {
	type surrogate Credentials
	payload := struct {
		AppSettings  surrogate `json:"boxAppSettings"`
		EnterpriseID string    `json:"enterpriseID"`
	}{
		AppSettings:  surrogate(*creds),
		EnterpriseID: creds.EnterpriseID,
	}
	data, err := json.Marshal(payload)
	return data, errors.JSONMarshalError.Wrap(err)
}

// UnmarshalJSON unmarshals JSON into this
func (creds *Credentials) UnmarshalJSON(payload []byte) (err error) {
	type surrogate Credentials
	var data = struct {
		AppSettings  surrogate `json:"boxAppSettings"`
		EnterpriseID string    `json:"enterpriseID"`
	}{}
	if err = json.Unmarshal(payload, &data); err != nil {
		return errors.JSONUnmarshalError.Wrap(err)
	}
	*creds = Credentials(data.AppSettings)
	creds.EnterpriseID = data.EnterpriseID
	return
}