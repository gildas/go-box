package box_test

import (
	"context"
	"encoding/json"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/gildas/go-box"
	"github.com/gildas/go-core"
	"github.com/gildas/go-errors"
	"github.com/gildas/go-logger"
	"github.com/gildas/go-request"
	"github.com/stretchr/testify/suite"
)

type SharedLinkSuite struct {
	suite.Suite
	Name   string
	Logger *logger.Logger
	Start  time.Time

	Client *box.Client
	Root   *box.FolderEntry
}

func TestSharedLinkSuite(t *testing.T) {
	suite.Run(t, new(SharedLinkSuite))
}

func (suite *SharedLinkSuite) SetupSuite() {
	suite.Name = strings.TrimSuffix(reflect.TypeOf(*suite).Name(), "Suite")
	folder := filepath.Join(".", "log")
	if err := os.MkdirAll(folder, os.ModePerm); err != nil {
		panic(err)
	}
	suite.Logger = logger.CreateWithStream("test", &logger.FileStream{Path: filepath.Join(folder, "test-"+strings.ToLower(suite.Name)+".log"), FilterLevel: logger.TRACE, Unbuffered: true})
}

func (suite *SharedLinkSuite) TearDownSuite() {
	folder, err := suite.Client.Folders.FindByName(context.Background(), "unit-test")
	if err == nil {
		err := suite.Client.Folders.Delete(context.Background(), folder)
		suite.Assert().Nilf(err, "Failed deleting root folder. Error: %s", err)
	}
}

func (suite *SharedLinkSuite) BeforeTest(suiteName, testName string) {
	var err error

	suite.Logger.Infof("Test Start: %s %s", testName, strings.Repeat("-", 80-13-len(testName)))
	suite.Start = time.Now()

	if suite.Client == nil {
		suite.Logger.Infof("Creating a new box.Client")
		suite.Client = box.NewClient(suite.Logger.ToContext(context.Background()))
	}
	if !suite.Client.IsAuthenticated() {
		err = suite.Client.Auth.Authenticate(context.Background(), suite.FetchCredentials())
		suite.Require().Nil(err, "Failed to authenticate box.Client")
	}

	if suite.Root == nil {
		suite.Root, err = suite.Client.Folders.FindByName(context.Background(), "unit-test")
		if err != nil {
			suite.Root, err = suite.Client.Folders.Create(context.Background(), &box.FolderEntry{
				Name: "unit-test",
			})
		}
		suite.Require().Nilf(err, "Failed creating root folder. Error: %s", err)
	}
}

func (suite *SharedLinkSuite) AfterTest(suiteName, testName string) {
	duration := time.Since(suite.Start)
	suite.Logger.Record("duration", duration.String()).Infof("Test End: %s %s", testName, strings.Repeat("-", 80-11-len(testName)))
	if suite.Root != nil {
		if !suite.Client.IsAuthenticated() {
			err := suite.Client.Auth.Authenticate(context.Background(), suite.FetchCredentials())
			suite.Require().Nil(err, "Failed to authenticate box.Client")
		}
		err := suite.Client.Folders.Delete(context.Background(), suite.Root)
		suite.Assert().Nilf(err, "Failed deleting root folder. Error: %s", err)
		suite.Root = nil
	}
}

func (suite *SharedLinkSuite) FetchCredentials() box.Credentials {
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

func (suite *SharedLinkSuite) TestCanMarshalSharedLinkOptions() {
	slo := box.SharedLinkOptions{
		Access: "read",
		Permissions: box.Permissions{
			CanPreview: true,
		},
	}

	payload, err := json.Marshal(slo)
	suite.Require().Nilf(err, "Error should be nil. Error: %v", err)
	suite.Assert().NotEmpty(payload)
}

func (suite *SharedLinkSuite) TestCanMarshalSharedLink() {
	slURL, _ := url.Parse("https://www.acme.org/files/1234")
	sl := box.SharedLink{
		URL:    slURL,
		Access: "read",
		Permissions: box.Permissions{
			CanPreview: true,
		},
	}

	payload, err := json.Marshal(sl)
	suite.Require().Nilf(err, "Error should be nil. Error: %v", err)
	suite.Assert().NotEmpty(payload)
}

func (suite *SharedLinkSuite) TestShouldFailUnmarshalSharedLinkWithInvalidJSON() {
	var sl box.SharedLink
	err := json.Unmarshal([]byte(`{"Access": 1234}`), &sl)
	suite.Require().NotNil(err, "Should have failed unmarshaling")
	suite.Assert().Truef(errors.Is(err, errors.JSONUnmarshalError), "Error should be an JSON Unmarshal Error. Error: %+v", err)

}

func (suite *SharedLinkSuite) TestCanCreateWithoutOptions() {
	uploaded, err := suite.Client.Files.Upload(context.Background(), &box.UploadOptions{
		Parent:   suite.Root.AsPathEntry(),
		Filename: "hello.txt",
		Content:  request.ContentWithData([]byte("Hello, World!"), "text/plain"),
	})
	suite.Require().Nilf(err, "Failed uploading a file. Error: %s", err)
	entry, err := suite.Client.Files.FindByID(context.Background(), uploaded.Files[0].ID)
	suite.Require().Nilf(err, "Failed finding a file. Error: %s", err)

	sl, err := suite.Client.SharedLinks.Create(context.Background(), entry, nil)
	suite.Require().Nilf(err, "Error should be nil. Error: %v", err)
	suite.Require().NotNil(sl, "SharedLink should not be nil")
	suite.Assert().NotNil(sl.URL, "SharedLink's URL should not be nil")
}

func (suite *SharedLinkSuite) TestCanCreateWithOptions() {
	uploaded, err := suite.Client.Files.Upload(context.Background(), &box.UploadOptions{
		Parent:   suite.Root.AsPathEntry(),
		Filename: "hello.txt",
		Content:  request.ContentWithData([]byte("Hello, World!"), "text/plain"),
	})
	suite.Require().Nilf(err, "Failed uploading a file. Error: %s", err)
	entry, err := suite.Client.Files.FindByID(context.Background(), uploaded.Files[0].ID)
	suite.Require().Nilf(err, "Failed finding a file. Error: %s", err)

	sl, err := suite.Client.SharedLinks.Create(context.Background(), entry, &box.SharedLinkOptions{Access: "open"})
	suite.Require().Nilf(err, "Error should be nil. Error: %v", err)
	suite.Require().NotNil(sl, "SharedLink should not be nil")
	suite.Assert().NotNil(sl.URL, "SharedLink's URL should not be nil")
}

func (suite *SharedLinkSuite) TestShouldFailCreatingWithMissingEntry() {
	_, err := suite.Client.SharedLinks.Create(context.Background(), nil, nil)
	suite.Require().NotNil(err, "Should have failed sharing link")
	suite.Assert().Truef(errors.Is(err, errors.ArgumentMissingError), "Errors should be an Argument Missing Error. Error: %v", err)
	var details *errors.Error
	suite.Require().True(errors.As(err, &details), "Error should be an errors.Error")
	suite.Assert().Equal("entry", details.What)
}

func (suite *SharedLinkSuite) TestShouldFailCreatingWithInvalidEntry() {
	entry := &box.FileEntry{}
	_, err := suite.Client.SharedLinks.Create(context.Background(), entry, nil)
	suite.Require().NotNil(err, "Should have failed sharing link")
	suite.Assert().Truef(errors.Is(err, errors.ArgumentMissingError), "Errors should be an Argument Missing Error. Error: %v", err)
	var details *errors.Error
	suite.Require().True(errors.As(err, &details), "Error should be an errors.Error")
	suite.Assert().Equal("id", details.What)
}

func (suite *SharedLinkSuite) TestShouldFailCreatingWhenNotAuthenticated() {
	if suite.Client.IsAuthenticated() {
		suite.Client.Auth.Token = nil
	}

	entry := &box.FileEntry{ID: "1234"}
	_, err := suite.Client.SharedLinks.Create(context.Background(), entry, nil)
	suite.Require().NotNil(err, "Should have failed sharing link")
	suite.Assert().Truef(errors.Is(err, errors.UnauthorizedError), "Errors should be an Unauthorized Error. Error: %v", err)
}