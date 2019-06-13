package box

import (
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"github.com/youmark/pkcs8"
)

type boxClaims struct {
	BoxSubType string `json:"box_sub_type"`
	jwt.StandardClaims
}

// Authenticate authenticates with the given credentials
// Currently only AppAuth is supported
func (client *Client) Authenticate(auth Auth) (err error) {
	log := client.Logger.Scope("authenticate").Child()

	if client.Token != nil {
		// TODO: Check if the token expired!
		return nil
	}

	jwtToken := jwt.NewWithClaims(jwt.GetSigningMethod("RS256"), boxClaims{
		"enterprise",
		jwt.StandardClaims {
			Audience: "https://api.box.com/oauth2/token",
			ExpiresAt: time.Now().Add(30*time.Second).Unix(),
			Id:        uuid.Must(uuid.NewRandom()).String(),
			Issuer:    auth.ClientID,
			Subject:   auth.EnterpriseID,
		},
	})
	jwtToken.Header["kid"] = auth.AppAuth.PublicKeyID

	pemBlock, _ := pem.Decode([]byte(auth.AppAuth.PrivateKey))
	rsaKey, err := pkcs8.ParsePKCS8PrivateKeyRSA(pemBlock.Bytes, []byte(auth.AppAuth.Passphrase))
	if err != nil {
		return fmt.Errorf("Failed to parse RSA Key: %s", err)
	}

	signedToken, err := jwtToken.SignedString(rsaKey)
	if err != nil {
		return fmt.Errorf("Failed to sign the token. Error: %s", err)
	}

	form := url.Values{}
	form.Add("grant_type",    "urn:ietf:params:oauth:grant-type:jwt-bearer")
	form.Add("client_id",     auth.ClientID)
	form.Add("client_secret", auth.ClientSecret)
	form.Add("assertion",     signedToken)

	req, _ := http.NewRequest("POST", "https://api.box.com/oauth2/token", strings.NewReader(form.Encode()))
	req.PostForm = form
	req.Header.Add("User-Agent",   "BOX Client v." + VERSION)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	httpclient := http.DefaultClient
	if client.Proxy != nil {
		httpclient.Transport = &http.Transport{Proxy: http.ProxyURL(client.Proxy)}
	}

	log.Debugf("HTTP POST %s", "https://api.box.com/oauth2/token")
	start    := time.Now()
	res, err := httpclient.Do(req)
	duration := time.Since(start)
	if err != nil {
		return fmt.Errorf("Failed to send request to Box API: %s", err)
	}
	defer res.Body.Close()

	resbody, err := ioutil.ReadAll(res.Body) // read the body no matter what
	if err != nil {
		return fmt.Errorf("Failed to read response body: %s", err)
	}
	log.Debugf("Response in %s\nproto: %s,\nstatus: %s,\nheaders: %#v", duration, res.Proto, res.Status, res.Header)
	log.Tracef("Response body: %s", string(resbody))

	if res.StatusCode >= 300 {
		return fmt.Errorf("HTTP Error: %s", res.Status)
	}

	token := Token{}
	if err = json.Unmarshal(resbody, &token); err != nil {
		return fmt.Errorf("Failed to decode response body: %s", err)
	}
	// TODO: if we get an invalid_grant: https://github.com/box/box-node-sdk/blob/1d51f676b1323135891a70d470e2c9de97be9437/lib/token-manager.js#L216
	client.Token = &token
	return
}