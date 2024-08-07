package box_test

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/gildas/go-box"
	"github.com/gildas/go-core"
	"github.com/gildas/go-errors"
	"github.com/gildas/go-logger"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/suite"
)

type ClientSuite struct {
	suite.Suite
	Name   string
	Logger *logger.Logger
	Start  time.Time
}

func TestClientSuite(t *testing.T) {
	suite.Run(t, new(ClientSuite))
}

// *****************************************************************************
// Suite Tools

func (suite *ClientSuite) SetupSuite() {
	_ = godotenv.Load()
	suite.Name = strings.TrimSuffix(reflect.TypeOf(suite).Elem().Name(), "Suite")
	suite.Logger = logger.Create("test",
		&logger.FileStream{
			Path:         fmt.Sprintf("./log/test-%s.log", strings.ToLower(suite.Name)),
			Unbuffered:   true,
			SourceInfo:   true,
			FilterLevels: logger.NewLevelSet(logger.TRACE),
		},
	).Child("test", "test")
	suite.Logger.Infof("Suite Start: %s %s", suite.Name, strings.Repeat("=", 80-14-len(suite.Name)))
}

func (suite *ClientSuite) TearDownSuite() {
	if suite.T().Failed() {
		suite.Logger.Warnf("At least one test failed, we are not cleaning")
		suite.T().Log("At least one test failed, we are not cleaning")
	} else {
		suite.Logger.Infof("All tests succeeded, we are cleaning")
	}
	suite.Logger.Infof("Suite End: %s %s", suite.Name, strings.Repeat("=", 80-12-len(suite.Name)))
	suite.Logger.Close()
}

func (suite *ClientSuite) BeforeTest(suiteName, testName string) {
	suite.Logger.Infof("Test Start: %s %s", testName, strings.Repeat("-", 80-13-len(testName)))
	suite.Start = time.Now()
}

func (suite *ClientSuite) AfterTest(suiteName, testName string) {
	duration := time.Since(suite.Start)
	suite.Logger.Record("duration", duration.String()).Infof("Test End: %s %s", testName, strings.Repeat("-", 80-11-len(testName)))
}

func (suite *ClientSuite) FetchCredentials() box.Credentials {
	suite.Logger.Infof("Fetching credentials from environment")
	var credentials box.Credentials

	config := core.GetEnvAsString("BOX_CONFIG", "")
	if len(config) > 0 {
		suite.Logger.Debugf("Found BOX_CONFIG")
		err := json.Unmarshal([]byte(config), &credentials)
		suite.Require().Nil(err, "Failed to unmarshal BOX_CONFIG")
	} else {
		credentials = box.Credentials{
			ClientID:     core.GetEnvAsString("BOX_CLIENTID", ""),
			ClientSecret: core.GetEnvAsString("BOX_CLIENTSECRET", ""),
			EnterpriseID: core.GetEnvAsString("BOX_ENTERPRISEID", ""),
			AppAuth: box.AppAuth{
				PublicKeyID: core.GetEnvAsString("BOX_PUBLICKEYID", ""),
				PrivateKey:  core.GetEnvAsString("BOX_PRIVATEKEY", ""),
				Passphrase:  core.GetEnvAsString("BOX_PASSPHRASE", ""),
			},
		}
	}
	return credentials
}

// *****************************************************************************

func (suite *ClientSuite) TestCanCreateClient() {
	client := box.NewClient(suite.Logger.ToContext(context.Background()))
	suite.Require().NotNil(client)
	suite.Assert().NotNil(client.Auth, "Auth component should not be nil")
	suite.Assert().NotNil(client.Files, "Files component should not be nil")
	suite.Assert().NotNil(client.Logger, "Logger component should not be nil")
	suite.Assert().NotNil(client.Folders, "Folders component should not be nil")
	suite.Assert().NotNil(client.SharedLinks, "SharedLinks component should not be nil")
	suite.Assert().False(client.IsAuthenticated(), "Client should not be authenticated when just created")
}

func (suite *ClientSuite) TestCanCreateClientWithoutLogger() {
	client := box.NewClient(context.Background())
	suite.Require().NotNil(client)
	suite.Assert().NotNil(client.Auth, "Auth component should not be nil")
	suite.Assert().NotNil(client.Files, "Files component should not be nil")
	suite.Assert().NotNil(client.Logger, "Logger component should not be nil")
	suite.Assert().NotNil(client.Folders, "Folders component should not be nil")
	suite.Assert().NotNil(client.SharedLinks, "SharedLinks component should not be nil")
	suite.Assert().False(client.IsAuthenticated(), "Client should not be authenticated when just created")
}

func (suite *ClientSuite) TestCanMarshalCredentials() {
	credentials := box.Credentials{
		ClientID:     "someclientid",
		ClientSecret: "somesecret",
		AppAuth: box.AppAuth{
			PublicKeyID: "deadbeef",
			PrivateKey:  "-----BEGIN ENCRYPTED PRIVATE KEY-----\nzzz=\n-----END ENCRYPTED PRIVATE KEY-----\n",
			Passphrase:  "somepassphrase",
		},
		EnterpriseID: "12345678",
	}
	payload, err := json.Marshal(&credentials)
	suite.Require().Nil(err, "Failed to marshal credentials")
	suite.Assert().NotEmpty(payload)
}

func (suite *ClientSuite) TestCanUnmarshalCredentials() {
	var credentials box.Credentials
	config := `{
		"boxAppSettings": {
		  "clientId": "someclientid",
		  "clientSecret": "somesecret",
		  "appAuth": {
			"publicKeyId": "deadbeef",
			"privateKey": "-----BEGIN ENCRYPTED PRIVATE KEY-----\nzzz=\n-----END ENCRYPTED PRIVATE KEY-----\n",
			"passphrase": "somepassprase"
		  }
		},
		"enterpriseId": "12345678"
	}`
	err := json.Unmarshal([]byte(config), &credentials)
	suite.Require().Nil(err, "Failed to unmarshal a box.Credentials")
	suite.Assert().Equal("someclientid", credentials.ClientID)
	suite.Assert().Equal("somesecret", credentials.ClientSecret)
	suite.Assert().Equal("12345678", credentials.EnterpriseID)
}

func (suite *ClientSuite) TestShouldFailUnmarshalingCredentialsWithInvalidJSON() {
	var credentials box.Credentials
	config := `{"enterpriseId": 8}`
	err := json.Unmarshal([]byte(config), &credentials)
	suite.Require().NotNil(err, "Should have failed unmarshaling")
	suite.Assert().Truef(errors.Is(err, errors.JSONUnmarshalError), "Error should be an JSON Unmarshal Error. Error: %+v", err)
}

func (suite *ClientSuite) TestCanUnmarshalToken() {
	var token box.Token
	payload := `{
		"token_type":    "bearer",
		"access_token":  "123456789deadbeef",
		"expires_in":    3915,
		"restricted_to": []
	}`
	err := json.Unmarshal([]byte(payload), &token)
	suite.Require().Nil(err, "Failed to unmarshal a box.Token")
	suite.Assert().Equal("Bearer", token.TokenType)
	suite.Assert().Equal("123456789deadbeef", token.AccessToken)
}

func (suite *ClientSuite) TestShouldFailUnmarshalingTokenWithInvalidJSON() {
	var token box.Token
	config := `{"access_token": 8}`
	err := json.Unmarshal([]byte(config), &token)
	suite.Require().NotNil(err, "Should have failed unmarshaling")
	suite.Assert().Truef(errors.Is(err, errors.JSONUnmarshalError), "Error should be an JSON Unmarshal Error. Error: %+v", err)
}

func (suite *ClientSuite) TestCanStoreTokenIntoContext() {
	token := box.Token{
		TokenType:   "Bearer",
		AccessToken: "123456789deadbeef",
		ExpiresOn:   time.Now().UTC().Add(3915 * time.Millisecond),
	}
	ctx := token.ToContext(context.Background())
	suite.Require().NotNil(ctx, "The context should not be nil")

	token.AccessToken = ""
	ctx = token.ToContext(context.Background())
	suite.Require().Equal(context.Background(), ctx)
}

func (suite *ClientSuite) TestCanRetrieveTokenFromContext() {
	token := box.Token{
		TokenType:   "Bearer",
		AccessToken: "123456789deadbeef",
		ExpiresOn:   time.Now().UTC().Add(3915 * time.Millisecond),
	}
	ctx := token.ToContext(context.Background())
	suite.Require().NotNil(ctx, "The context should not be nil")

	stored := box.TokenFromContext(ctx)
	suite.Require().NotNil(stored, "The context should not be nil")
	suite.Assert().Equal(token.AccessToken, stored.AccessToken)
}

func (suite *ClientSuite) TestCanAuthenticate() {
	client := box.NewClient(suite.Logger.ToContext(context.Background()))
	suite.Require().NotNil(client)
	err := client.Auth.Authenticate(context.Background(), suite.FetchCredentials())
	suite.Assert().Nil(err, "Failed to authenticate box.Client")
	if err != nil {
		suite.Logger.Errorf("Failed authenticating", err)
	}
}

func (suite *ClientSuite) TestCanAuthenticateTwice() {
	client := box.NewClient(suite.Logger.ToContext(context.Background()))
	suite.Require().NotNil(client)
	err := client.Auth.Authenticate(context.Background(), suite.FetchCredentials())
	suite.Require().Nil(err, "Failed to authenticate box.Client")
	err = client.Auth.Authenticate(context.Background(), suite.FetchCredentials())
	suite.Assert().Nil(err, "Failed to authenticate box.Client again")
	if err != nil {
		suite.Logger.Errorf("Failed authenticating", err)
	}
}

func (suite *ClientSuite) TestShouldFailAuthenticatingWithMissingPrivateKey() {
	client := box.NewClient(suite.Logger.ToContext(context.Background()))
	suite.Require().NotNil(client)
	credentials := suite.FetchCredentials()
	credentials.AppAuth.PrivateKey = ""
	err := client.Auth.Authenticate(context.Background(), credentials)
	suite.Require().NotNil(err, "Should have Failed to authenticate box.Client")
	suite.Assert().True(errors.Is(err, errors.Unauthorized), "Error should be an Unauthorized Error")
	suite.Assert().True(errors.Is(err, box.InvalidPrivateKey), "Error should be an Invalid Private Key Error")
	suite.Logger.Errorf("Expected error", err)
}

func (suite *ClientSuite) TestShouldFailAuthenticatingWithInvalidPrivateKey() {
	client := box.NewClient(suite.Logger.ToContext(context.Background()))
	suite.Require().NotNil(client)
	credentials := suite.FetchCredentials()
	credentials.AppAuth.PrivateKey = "-----BEGIN ENCRYPTED PRIVATE KEY-----\nzzz=\n-----END ENCRYPTED PRIVATE KEY-----\n"
	err := client.Auth.Authenticate(context.Background(), credentials)
	suite.Require().NotNil(err, "Should have Failed to authenticate box.Client")
	suite.Logger.Errorf("(Expected) Failed to Authenticate", err)
	suite.Assert().Truef(errors.Is(err, errors.Unauthorized), "Error should be an Unauthorized Error, Error: %s", err)
	suite.Assert().Truef(errors.Is(err, box.InvalidPrivateKey), "Error should be an Invalid Private Key Error. Error: %s", err)
	suite.Logger.Errorf("Expected error", err)
}

func (suite *ClientSuite) TestShouldFailAuthenticatingWithInvalidClientID() {
	client := box.NewClient(suite.Logger.ToContext(context.Background()))
	suite.Require().NotNil(client)
	credentials := suite.FetchCredentials()
	credentials.ClientID = "Invalid ClientId"
	err := client.Auth.Authenticate(context.Background(), credentials)
	suite.Require().NotNil(err, "Should have Failed to authenticate box.Client")
	suite.Assert().True(errors.Is(err, errors.Unauthorized), "Error should be an Unauthorized Error. Error: %s", err)
	suite.Assert().True(errors.Is(err, box.InvalidGrant), "Error should be an Invalid Grant Error. Error: %v", err)
	suite.Logger.Errorf("Expected error", err)
}
