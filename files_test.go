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
	"github.com/gildas/go-request"
	"github.com/stretchr/testify/suite"
)

type FileSuite struct {
	suite.Suite
	Name   string
	Logger *logger.Logger
	Start  time.Time

	Client *box.Client
	Root   *box.FolderEntry
}

func TestFileSuite(t *testing.T) {
	suite.Run(t, new(FileSuite))
}

func (suite *FileSuite) SetupSuite() {
	suite.Name = strings.TrimSuffix(reflect.TypeOf(*suite).Name(), "Suite")
	folder := filepath.Join(".", "log")
	if err := os.MkdirAll(folder, os.ModePerm); err != nil {
		panic(err)
	}
	suite.Logger = logger.CreateWithStream("test", &logger.FileStream{Path: filepath.Join(folder, "test-"+strings.ToLower(suite.Name)+".log"), FilterLevel: logger.TRACE, Unbuffered: true})
}

func (suite *FileSuite) TearDownSuite() {
	folder, err := suite.Client.Folders.FindByName(context.Background(), "unit-test")
	if err == nil {
		err := suite.Client.Folders.Delete(context.Background(), folder)
		suite.Assert().Nilf(err, "Failed deleting root folder. Error: %s", err)
	}
}

func (suite *FileSuite) BeforeTest(suiteName, testName string) {
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

func (suite *FileSuite) AfterTest(suiteName, testName string) {
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

func (suite *FileSuite) FetchCredentials() box.Credentials {
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

func (suite *FileSuite) TestCanDownload() {
	collection, err := suite.Client.Files.Upload(context.Background(), &box.UploadOptions{
		Parent:   suite.Root.AsPathEntry(),
		Filename: "hello.txt",
		Content:  request.ContentWithData([]byte("Hello, World!"), "text/plain"),
	})
	suite.Require().Nilf(err, "Failed uploading a file. Error: %s", err)
	suite.Require().NotNil(collection, "File Collection should not be nil")
	suite.Require().Equal(1, collection.Count, "There should be 1 entry in the collection")
	entry := collection.Files[0]
	suite.Require().NotNil(entry, "The first entry in the collection should not be nil")
	suite.Assert().Equal("file", entry.Type)
	suite.Assert().Equal("hello.txt", entry.Name)

	downloaded, err := suite.Client.Files.Download(context.Background(), &entry)
	suite.Require().Nilf(err, "Failed downloading a file. Error: %s", err)
	suite.Require().NotNil(downloaded, "Content should not be nil")
	suite.Assert().Equal("text/plain", downloaded.Type)
	suite.Assert().Equal(int64(13), downloaded.Length)
}

func (suite *FileSuite) TestCanUploadWithPayload() {
	collection, err := suite.Client.Files.Upload(context.Background(), &box.UploadOptions{
		Parent:   suite.Root.AsPathEntry(),
		Filename: "hello.json",
		Payload: struct {
			ID      string `json:"id"`
			Message string `json:"message"`
		}{"1234", "Hello, World"},
	})
	suite.Require().Nilf(err, "Failed uploading a file. Error: %s", err)
	suite.Require().NotNil(collection, "File Collection should not be nil")
	suite.Require().Equal(1, collection.Count, "There should be 1 entry in the collection")
	entry := collection.Files[0]
	suite.Require().NotNil(entry, "The first entry in the collection should not be nil")
	suite.Assert().Equal("file", entry.Type)
	suite.Assert().Equal("hello.json", entry.Name)
}

func (suite *FileSuite) TestCanUploadWithContent() {
	collection, err := suite.Client.Files.Upload(context.Background(), &box.UploadOptions{
		Parent:   suite.Root.AsPathEntry(),
		Filename: "hello.txt",
		Content:  request.ContentWithData([]byte("Hello, World!"), "text/plain"),
	})
	suite.Require().Nilf(err, "Failed uploading a file. Error: %s", err)
	suite.Require().NotNil(collection, "File Collection should not be nil")
	suite.Require().Equal(1, collection.Count, "There should be 1 entry in the collection")
	entry := collection.Files[0]
	suite.Require().NotNil(entry, "The first entry in the collection should not be nil")
	suite.Assert().Equal("file", entry.Type)
	suite.Assert().Equal("hello.txt", entry.Name)
}

func (suite *FileSuite) TestCanFindByID() {
	uploaded, err := suite.Client.Files.Upload(context.Background(), &box.UploadOptions{
		Parent:   suite.Root.AsPathEntry(),
		Filename: "hello.txt",
		Content:  request.ContentWithData([]byte("Hello, World!"), "text/plain"),
	})
	suite.Require().Nilf(err, "Failed uploading a file. Error: %s", err)
	entry, err := suite.Client.Files.FindByID(context.Background(), uploaded.Files[0].ID)
	suite.Require().Nilf(err, "Failed finding a file. Error: %s", err)
	suite.Require().NotNil(entry, "File Entry should not be nil")
	suite.Assert().Equal("file", entry.Type)
	suite.Assert().Equal("hello.txt", entry.Name)
}

func (suite *FileSuite) TestCanFindByName() {
	_, err := suite.Client.Files.Upload(context.Background(), &box.UploadOptions{
		Parent:   suite.Root.AsPathEntry(),
		Filename: "hello.txt",
		Content:  request.ContentWithData([]byte("Hello, World!"), "text/plain"),
	})
	suite.Require().Nilf(err, "Failed uploading a file. Error: %s", err)
	entry, err := suite.Client.Files.FindByName(context.Background(), "hello.txt", suite.Root.AsPathEntry())
	suite.Require().Nilf(err, "Failed finding a file. Error: %s", err)
	suite.Require().NotNil(entry, "File Entry should not be nil")
	suite.Assert().Equal("file", entry.Type)
	suite.Assert().Equal("hello.txt", entry.Name)
}

func (suite *FileSuite) TestCanMarshalFileEntry() {
	entry := &box.FileEntry{}
	payload, err := json.Marshal(entry)
	suite.Require().Nilf(err, "Failed marshaling a FileEntry. Error: %s", err)
	suite.Require().NotEmpty(payload, "Payload should not be nil")
}

func (suite *FileSuite) TestShouldFailDownloadingWithMissingEntry() {
	_, err := suite.Client.Files.Download(context.Background(), nil)
	suite.Require().NotNil(err, "Should have failed downloading file")
	suite.Assert().Truef(errors.Is(err, errors.ArgumentMissing), "Errors should be an Argument Missing Error. Error: %v", err)
	var details *errors.Error
	suite.Require().True(errors.As(err, &details), "Error should be an errors.Error")
	suite.Assert().Equal("entry", details.What)
}

func (suite *FileSuite) TestShouldFailDownloadingWhenNotAuthenticated() {
	if suite.Client.IsAuthenticated() {
		suite.Client.Auth.Token = nil
	}
	_, err := suite.Client.Files.Download(context.Background(), &box.FileEntry{ID: "1234"})
	suite.Require().NotNil(err, "Should have failed dowloading file")
	suite.Assert().Truef(errors.Is(err, errors.Unauthorized), "Errors should be an Unauthorized Error. Error: %v", err)
}

func (suite *FileSuite) TestShouldFailFindingWithMissingParent() {
	_, err := suite.Client.Files.FindByName(context.Background(), "hello.txt", nil)
	suite.Require().NotNil(err, "Should have failed finding FileEntry")
	suite.Assert().Truef(errors.Is(err, errors.ArgumentMissing), "Errors should be an Argument Missing Error. Error: %v", err)
	var details *errors.Error
	suite.Require().True(errors.As(err, &details), "Error should be an errors.Error")
	suite.Assert().Equal("parent", details.What)
}

func (suite *FileSuite) TestShouldFailFindingWithMissingID() {
	_, err := suite.Client.Files.FindByID(context.Background(), "")
	suite.Require().NotNil(err, "Should have failed finding FileEntry")
	suite.Assert().Truef(errors.Is(err, errors.ArgumentMissing), "Errors should be an Argument Missing Error. Error: %v", err)
	var details *errors.Error
	suite.Require().True(errors.As(err, &details), "Error should be an errors.Error")
	suite.Assert().Equal("id", details.What)
}

func (suite *FileSuite) TestShouldFailFindingWithMissingFilename() {
	_, err := suite.Client.Files.FindByName(context.Background(), "", suite.Root.AsPathEntry())
	suite.Require().NotNil(err, "Should have failed finding FileEntry")
	suite.Assert().Truef(errors.Is(err, errors.ArgumentMissing), "Errors should be an Argument Missing Error. Error: %v", err)
	var details *errors.Error
	suite.Require().True(errors.As(err, &details), "Error should be an errors.Error")
	suite.Assert().Equal("filename", details.What)
}

func (suite *FileSuite) TestShouldFailFindingWithInvalidParent() {
	folder := &box.FolderEntry{Type: "folder", ID: "1234", Name: "bogus_folder"}
	_, err := suite.Client.Files.FindByName(context.Background(), "hello.txt", folder.AsPathEntry())
	suite.Require().NotNil(err, "Should have failed finding FileEntry")
	suite.Assert().Truef(errors.Is(err, errors.ArgumentInvalid), "Errors should be an Argument Invalid Error. Error: %v", err)
	var details *errors.Error
	suite.Require().True(errors.As(err, &details), "Error should be an errors.Error")
	suite.Assert().Equal("parent", details.What)
	suite.Assert().Equal("1234", details.Value.(string))
}

func (suite *FileSuite) TestShouldFailFindingWithUnknownFilename() {
	_, err := suite.Client.Files.FindByName(context.Background(), "this_is_not_the_file_you_are_looking_for.txt", suite.Root.AsPathEntry())
	suite.Require().NotNil(err, "Should have failed finding FileEntry")
	suite.Assert().Truef(errors.Is(err, errors.NotFound), "Errors should be a Not Found Error. Error: %v", err)
	var details *errors.Error
	suite.Require().True(errors.As(err, &details), "Error should be an errors.Error")
	suite.Assert().Equal("filename", details.What)
	suite.Assert().Equal("this_is_not_the_file_you_are_looking_for.txt", details.Value.(string))
}

func (suite *FileSuite) TestShouldFailFindingWhenNotAuthenticated() {
	if suite.Client.IsAuthenticated() {
		suite.Client.Auth.Token = nil
	}
	_, err := suite.Client.Files.FindByID(context.Background(), "12345")
	suite.Require().NotNil(err, "Should have failed finding FileEntry")
	suite.Assert().Truef(errors.Is(err, errors.Unauthorized), "Errors should be an Unauthorized Error. Error: %v", err)

	_, err = suite.Client.Files.FindByName(context.Background(), "hello.txt", suite.Root.AsPathEntry())
	suite.Require().NotNil(err, "Should have failed finding FileEntry")
	suite.Assert().Truef(errors.Is(err, errors.Unauthorized), "Errors should be an Unauthorized Error. Error: %v", err)
}

func (suite *FileSuite) TestShouldFailUnmarshalingFileEntryWithInvalidJSON() {
	var entry box.FileEntry
	err := json.Unmarshal([]byte(`{"Type": "file", "id": 1234}`), &entry)
	suite.Require().NotNil(err, "Should have failed unmarshaling")
	suite.Assert().Truef(errors.Is(err, errors.JSONUnmarshalError), "Error should be an JSON Unmarshal Error. Error: %+v", err)
}

func (suite *FileSuite) TestShouldFailUploadingWithoutOptions() {
	_, err := suite.Client.Files.Upload(context.Background(), nil)
	suite.Require().NotNil(err, "Should have failed uploading file")
	suite.Assert().Truef(errors.Is(err, errors.ArgumentMissing), "Errors should be an Argument Missing Error. Error: %v", err)
	var details *errors.Error
	suite.Require().True(errors.As(err, &details), "Error should be an errors.Error")
	suite.Assert().Equal("options", details.What)
}

func (suite *FileSuite) TestShouldFailUploadingWithoutFilename() {
	_, err := suite.Client.Files.Upload(context.Background(), &box.UploadOptions{})
	suite.Require().NotNil(err, "Should have failed uploading file")
	suite.Assert().Truef(errors.Is(err, errors.ArgumentMissing), "Errors should be an Argument Missing Error. Error: %v", err)
	var details *errors.Error
	suite.Require().True(errors.As(err, &details), "Error should be an errors.Error")
	suite.Assert().Equal("filename", details.What)
}

func (suite *FileSuite) TestShouldFailUploadingWithoutContent() {
	_, err := suite.Client.Files.Upload(context.Background(), &box.UploadOptions{
		Parent:   suite.Root.AsPathEntry(),
		Filename: "hello.txt",
	})
	suite.Require().NotNil(err, "Should have failed uploading file")
	suite.Assert().Truef(errors.Is(err, errors.ArgumentMissing), "Errors should be an Argument Missing Error. Error: %v", err)
	var details *errors.Error
	suite.Require().True(errors.As(err, &details), "Error should be an errors.Error")
	suite.Assert().Equal("content", details.What)
}

func (suite *FileSuite) TestShouldFailUploadingWithBogusPayload() {
	_, err := suite.Client.Files.Upload(context.Background(), &box.UploadOptions{
		Parent:   suite.Root.AsPathEntry(),
		Filename: "hello.txt",
		Payload:  BogusData{ID: "1234"},
	})
	suite.Require().NotNil(err, "Should have failed uploading file")
	suite.Assert().Truef(errors.Is(err, errors.JSONMarshalError), "Errors should be a JSON Marshal Error. Error: %v", err)
}

func (suite *FileSuite) TestShouldFailUploadingWithInvalidParent() {
	folder := &box.FolderEntry{Type: "folder", ID: "1234", Name: "bogus_folder"}
	_, err := suite.Client.Files.Upload(context.Background(), &box.UploadOptions{
		Parent:   folder.AsPathEntry(),
		Filename: "hello.txt",
		Content:  request.ContentWithData([]byte("Hello, World!"), "text/plain"),
	})
	suite.Require().NotNil(err, "Should have failed uploading file")
	suite.Assert().Truef(errors.Is(err, errors.NotFound), "Errors should be a Not Found Error. Error: %v", err)
}

func (suite *FileSuite) TestShouldFailUploadingWhenNotAuthenticated() {
	if suite.Client.IsAuthenticated() {
		suite.Client.Auth.Token = nil
	}
	_, err := suite.Client.Files.Upload(context.Background(), &box.UploadOptions{
		Parent:   suite.Root.AsPathEntry(),
		Filename: "hello.txt",
		Content:  request.ContentWithData([]byte("Hello, World!"), "text/plain"),
	})
	suite.Require().NotNil(err, "Should have failed uploading file")
	suite.Assert().Truef(errors.Is(err, errors.Unauthorized), "Errors should be an Unauthorized Error. Error: %v", err)
}
