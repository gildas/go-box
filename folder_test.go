package box_test

import (
	"context"
	"encoding/json"
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
	"github.com/stretchr/testify/suite"
)

type FolderSuite struct {
	suite.Suite
	Name   string
	Logger *logger.Logger
	Start  time.Time

	Client *box.Client
	Root   *box.FolderEntry
}

func TestFolderSuite(t *testing.T) {
	suite.Run(t, new(FolderSuite))
}

func (suite *FolderSuite) SetupSuite() {
	suite.Name = strings.TrimSuffix(reflect.TypeOf(*suite).Name(), "Suite")
	folder := filepath.Join(".", "log")
	if err := os.MkdirAll(folder, os.ModePerm); err != nil {
		panic(err)
	}
	suite.Logger = logger.CreateWithStream("test", &logger.FileStream{Path: filepath.Join(folder, "test-"+strings.ToLower(suite.Name)+".log"), FilterLevel: logger.TRACE, Unbuffered: true})
}

func (suite *FolderSuite) TearDownSuite() {
	folder, err := suite.Client.Folders.FindByName(context.Background(), "unit-test")
	if err == nil {
		err := suite.Client.Folders.Delete(context.Background(), folder)
		suite.Assert().Nilf(err, "Failed deleting root folder. Error: %s", err)
	}
}

func (suite *FolderSuite) BeforeTest(suiteName, testName string) {
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

func (suite *FolderSuite) AfterTest(suiteName, testName string) {
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

func (suite *FolderSuite) FetchCredentials() box.Credentials {
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

func (suite *FolderSuite) TestCanCreate() {
	// The root folder is created in the "BeforeTest" func, we just need to test it
	suite.Require().NotNil(suite.Root, "Folder entry should not be nil")
	suite.Assert().Equal("unit-test", suite.Root.Name)
	suite.Assert().NotEqual("0", suite.Root.ID)
	suite.Assert().Equal("0", suite.Root.Parent.ID)
	suite.Assert().Equal("active", suite.Root.ItemStatus)
	suite.Logger.Record("root", suite.Root).Infof("Root folder created")
}

func (suite *FolderSuite) TestCanDelete() {
	if suite.Root != nil {
		err := suite.Client.Folders.Delete(context.Background(), suite.Root)
		suite.Assert().Nilf(err, "Failed deleting root folder. Error: %s", err)
		suite.Root = nil
	}
}

func (suite *FolderSuite) TestCanCreateSubFolder() {
	sub, err := suite.Client.Folders.Create(context.Background(), &box.FolderEntry{
		Name:   "subfolder",
		Parent: suite.Root.AsPathEntry(),
	})
	suite.Require().Nilf(err, "Failed creating a folder. Error: %s", err)
	suite.Require().NotNil(sub, "Subfolder entry should not be nil")
}

func (suite *FolderSuite) TestCanFindByID() {
	folder, err := suite.Client.Folders.FindByID(context.Background(), suite.Root.ID)
	suite.Require().Nilf(err, "Failed finding a folder. Error: %s", err)
	suite.Assert().Equal(suite.Root, folder)
}

func (suite *FolderSuite) TestCanFindByName() {
	folder, err := suite.Client.Folders.FindByName(context.Background(), suite.Root.Name)
	suite.Require().Nilf(err, "Failed finding a folder. Error: %s", err)
	suite.Assert().Equal(suite.Root, folder)
}

func (suite *FolderSuite) TestShouldFailCreatingWithMissingName() {
	_, err := suite.Client.Folders.Create(context.Background(), &box.FolderEntry{})
	suite.Require().NotNil(err, "Should have failed creating folder")
	suite.Assert().Truef(errors.Is(err, errors.ArgumentMissingError), "Errors should be an Argument Missing Error. Error: %v", err)
	var details *errors.Error
	suite.Require().True(errors.As(err, &details), "Error should be an errors.Error")
	suite.Assert().Equal("name", details.What)
}

func (suite *FolderSuite) TestShouldFailCreatingWhenNotAuthenticated() {
	if suite.Client.IsAuthenticated() {
		suite.Client.Auth.Token = nil
	}
	_, err := suite.Client.Folders.Create(context.Background(), &box.FolderEntry{
		Name: "unit-test",
	})
	suite.Require().NotNil(err, "Should have failed creating folder")
	suite.Assert().Truef(errors.Is(err, errors.UnauthorizedError), "Errors should be an Unauthorized Error. Error: %v", err)
}

func (suite *FolderSuite) TestShouldFailCreatingWithSameName() {
	_, err := suite.Client.Folders.Create(context.Background(), &box.FolderEntry{
		Name: "unit-test",
	})
	suite.Require().NotNil(err, "Should have failed creating folder")
	suite.Assert().Truef(errors.Is(err, box.ItemNameInUseError), "Errors should be an Item Name In Use Error. Error: %v", err)
}

func (suite *FolderSuite) TestShouldFailDeletingWithInvalidID() {
	folder := *suite.Root
	folder.ID = "1234"
	err := suite.Client.Folders.Delete(context.Background(), &folder)
	suite.Require().NotNil(err, "Should have failed deleting folder with invalid ID")
	suite.Assert().Truef(errors.Is(err, errors.NotFoundError), "Errors should be a Not Found Error. Error: %v", err)
}

func (suite *FolderSuite) TestShouldFailDeletingWithMissingID() {
	folder := *suite.Root
	folder.ID = ""
	err := suite.Client.Folders.Delete(context.Background(), &folder)
	suite.Require().NotNil(err, "Should have failed deleting folder without ID")
	suite.Assert().Truef(errors.Is(err, errors.ArgumentMissingError), "Errors should be an Argument Missing Error. Error: %v", err)
	var details *errors.Error
	suite.Require().True(errors.As(err, &details), "Error should be an errors.Error")
	suite.Assert().Equal("ID", details.What)
}

func (suite *FolderSuite) TestShouldFailDeletingWhenNotAuthenticated() {
	if suite.Client.IsAuthenticated() {
		suite.Client.Auth.Token = nil
	}
	err := suite.Client.Folders.Delete(context.Background(), suite.Root)
	suite.Require().NotNil(err, "Should have failed deleting folder when unauthenticated")
	suite.Assert().Truef(errors.Is(err, errors.UnauthorizedError), "Errors should be an Unauthorized Error. Error: %v", err)
}

func (suite *FolderSuite) TestShouldFailFindingWithInvalidID() {
	_, err := suite.Client.Folders.FindByID(context.Background(), "1234")
	suite.Require().NotNil(err, "Should have failed finding folder with invalid ID")
	suite.Assert().Truef(errors.Is(err, errors.NotFoundError), "Errors should be a Not Found Error. Error: %v", err)
}

func (suite *FolderSuite) TestShouldFailFindingWithInvalidName() {
	_, err := suite.Client.Folders.FindByName(context.Background(), "this_is_not_the_folder_you_are_looking_for")
	suite.Require().NotNil(err, "Should have failed finding folder with invalid name")
	suite.Assert().Truef(errors.Is(err, errors.NotFoundError), "Errors should be a Not Found Error. Error: %v", err)
	var details *errors.Error
	suite.Require().True(errors.As(err, &details), "Error should be an errors.Error")
	suite.Assert().Equal("folder", details.What)
	suite.Assert().Equal("this_is_not_the_folder_you_are_looking_for", details.Value.(string))
}

func (suite *FolderSuite) TestShouldFailFindingWithMissingID() {
	_, err := suite.Client.Folders.FindByID(context.Background(), "")
	suite.Require().NotNil(err, "Should have failed finding folder without ID")
	suite.Assert().Truef(errors.Is(err, errors.ArgumentMissingError), "Errors should be an Argument Missing Error. Error: %v", err)
	var details *errors.Error
	suite.Require().True(errors.As(err, &details), "Error should be an errors.Error")
	suite.Assert().Equal("ID", details.What)
}

func (suite *FolderSuite) TestShouldFailFindingWithMissingName() {
	_, err := suite.Client.Folders.FindByName(context.Background(), "")
	suite.Require().NotNil(err, "Should have failed finding folder without name")
	suite.Assert().Truef(errors.Is(err, errors.ArgumentMissingError), "Errors should be an Argument Missing Error. Error: %v", err)
	var details *errors.Error
	suite.Require().True(errors.As(err, &details), "Error should be an errors.Error")
	suite.Assert().Equal("name", details.What)
}

func (suite *FolderSuite) TestShouldFailFindingWhenNotAuthenticated() {
	if suite.Client.IsAuthenticated() {
		suite.Client.Auth.Token = nil
	}
	_, err := suite.Client.Folders.FindByID(context.Background(), suite.Root.ID)
	suite.Require().NotNil(err, "Should have failed deleting folder when unauthenticated")
	suite.Assert().Truef(errors.Is(err, errors.UnauthorizedError), "Errors should be an Unauthorized Error. Error: %v", err)

	_, err = suite.Client.Folders.FindByName(context.Background(), suite.Root.Name)
	suite.Require().NotNil(err, "Should have failed deleting folder when unauthenticated")
	suite.Assert().Truef(errors.Is(err, errors.UnauthorizedError), "Errors should be an Unauthorized Error. Error: %v", err)
}

func (suite *FolderSuite) TestShouldFailUnmarshalingFolderWithInvalidJSON() {
	var folder box.FolderEntry
	config := `{"id": 8}`
	err := json.Unmarshal([]byte(config), &folder)
	suite.Require().NotNil(err, "Should have failed unmarshaling")
	suite.Assert().Truef(errors.Is(err, errors.JSONUnmarshalError), "Error should be an JSON Unmarshal Error. Error: %+v", err)
}