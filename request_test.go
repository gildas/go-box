package box

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/gildas/go-errors"
	"github.com/gildas/go-logger"
	"github.com/gildas/go-request"
	"github.com/stretchr/testify/suite"
)

type RequestSuite struct {
	suite.Suite
	Name   string
	Logger *logger.Logger
	Start  time.Time

	Server    *httptest.Server
	ServerURL *url.URL
}

func TestRequestSuite(t *testing.T) {
	suite.Run(t, new(RequestSuite))
}

func (suite *RequestSuite) CreateServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		// Verify expected headers
		suite.Assert().Equal("BOX Client "+VERSION, req.Header.Get("User-Agent"))
		suite.Assert().NotEmpty(req.Header.Get("X-Request-Id"))

		switch req.Method {
		case http.MethodGet:
			switch req.URL.Path {
			case "/details/forbidden":
				res.Header().Set("Content-Type", "application/json")
				res.WriteHeader(http.StatusForbidden)
				payload, _ := json.Marshal(Forbidden)
				_, _ = res.Write(payload)
			case "/details/emptygrant":
				res.Header().Set("Content-Type", "application/json")
				res.WriteHeader(http.StatusBadRequest)
				payload, _ := json.Marshal(InvalidGrant)
				_, _ = res.Write(payload)
			case "/details/notfound":
				res.Header().Set("Content-Type", "application/json")
				res.WriteHeader(http.StatusNotFound)
				payload, _ := json.Marshal(FolderNotEmpty)
				_, _ = res.Write(payload)
			case "/details/unauthorized":
				res.Header().Set("Content-Type", "application/json")
				res.WriteHeader(http.StatusUnauthorized)
				payload, _ := json.Marshal(InvalidPrivateKey)
				_, _ = res.Write(payload)
			case "/unauthorized":
				res.Header().Set("Content-Type", "text/plain")
				res.WriteHeader(http.StatusUnauthorized)
				_, _ = res.Write([]byte("unauthorized"))
			case "/notfound":
				res.Header().Set("Content-Type", "text/plain")
				res.WriteHeader(http.StatusNotFound)
				_, _ = res.Write([]byte("not found"))
			default:
				res.Header().Set("Content-Type", "text/plain")
				res.WriteHeader(http.StatusNotFound)
				_, _ = res.Write([]byte("not found"))
			}
		default:
			res.Header().Set("Content-Type", "text/plain")
			res.WriteHeader(http.StatusMethodNotAllowed)
			_, _ = res.Write([]byte("not allowed"))
		}
	}))
}

func (suite *RequestSuite) TestShouldFailSendingWithoutOptions() {
	client := NewClient(suite.Logger.ToContext(context.Background()))
	suite.Require().NotNil(client)
	_, err := client.sendRequest(context.Background(), nil, nil)
	suite.Require().NotNil(err, "Should have failed sending request")
	suite.Assert().Truef(errors.Is(err, errors.ArgumentMissing), "Errors should be an Argument Missing Error. Error: %v", err)
	var details *errors.Error
	suite.Require().True(errors.As(err, &details), "Error should be an errors.Error")
	suite.Assert().Equal("options", details.What)
}

func (suite *RequestSuite) TestShouldReceiveUnauthorizedError() {
	reqURL, _ := suite.ServerURL.Parse("/unauthorized")
	client := NewClient(suite.Logger.ToContext(context.Background()))
	suite.Require().NotNil(client)
	_, err := client.sendRequest(context.Background(), &request.Options{
		URL:    reqURL,
		Logger: suite.Logger,
	}, nil)
	suite.Require().NotNil(err, "Should have failed sending request")
	suite.Logger.Errorf("Analyzing Error", err)
	suite.Assert().Truef(errors.Is(err, errors.Unauthorized), "Errors should be an Unauthorized Error. Error: %v", err)
	var details *RequestError
	suite.Assert().False(errors.As(err, &details), "Error should not contain a RequestError")
	if details != nil {
		suite.Logger.Errorf("Analyzing Details", details)
	}
}

func (suite *RequestSuite) TestShouldReceiveNotFoundError() {
	reqURL, _ := suite.ServerURL.Parse("/notfound")
	client := NewClient(suite.Logger.ToContext(context.Background()))
	suite.Require().NotNil(client)
	_, err := client.sendRequest(context.Background(), &request.Options{
		URL:    reqURL,
		Logger: suite.Logger,
	}, nil)
	suite.Require().NotNil(err, "Should have failed sending request")
	suite.Logger.Errorf("Analyzing Error", err)
	suite.Assert().Truef(errors.Is(err, errors.NotFound), "Errors should be a Not Found Error. Error: %v", err)
	var details *RequestError
	suite.Assert().False(errors.As(err, &details), "Error should not contain a RequestError")
	if details != nil {
		suite.Logger.Errorf("Analyzing Details", details)
	}
}

func (suite *RequestSuite) TestShouldReceiveNotAllowed() {
	reqURL, _ := suite.ServerURL.Parse("/this_is_not_the_page_you_are_looking_for")
	client := NewClient(suite.Logger.ToContext(context.Background()))
	suite.Require().NotNil(client)
	_, err := client.sendRequest(context.Background(), &request.Options{
		Method: http.MethodPost,
		URL:    reqURL,
		Logger: suite.Logger,
	}, nil)
	suite.Require().NotNil(err, "Should have failed sending request")
	suite.Logger.Errorf("Analyzing Error", err)
	suite.Assert().Truef(errors.Is(err, errors.HTTPMethodNotAllowed), "Errors should be an HTTP Method Not Allowed Error. Error: %v", err)
	var details *RequestError
	suite.Assert().False(errors.As(err, &details), "Error should not contain a RequestError")
	if details != nil {
		suite.Logger.Errorf("Analyzing Details", details)
	}
}

func (suite *RequestSuite) TestShouldReceiveError() {
	reqURL, _ := suite.ServerURL.Parse("/this_is_not_the_page_you_are_looking_for")
	client := NewClient(suite.Logger.ToContext(context.Background()))
	suite.Require().NotNil(client)
	_, err := client.sendRequest(context.Background(), &request.Options{
		URL:    reqURL,
		Logger: suite.Logger,
	}, nil)
	suite.Require().NotNil(err, "Should have failed sending request")
	suite.Logger.Errorf("Analyzing Error", err)
	suite.Assert().Truef(errors.Is(err, errors.HTTPNotFound), "Errors should be an HTTP Not Found Error. Error: %v", err)
	var details *RequestError
	suite.Assert().False(errors.As(err, &details), "Error should not contain a RequestError")
	if details != nil {
		suite.Logger.Errorf("Analyzing Details", details)
	}
}

func (suite *RequestSuite) TestShouldReceiveUnauthorizedErrorWithDetails() {
	reqURL, _ := suite.ServerURL.Parse("/details/unauthorized")
	client := NewClient(suite.Logger.ToContext(context.Background()))
	suite.Require().NotNil(client)
	_, err := client.sendRequest(context.Background(), &request.Options{
		URL:    reqURL,
		Logger: suite.Logger,
	}, nil)
	suite.Require().NotNil(err, "Should have failed sending request")
	suite.Logger.Errorf("Analyzing Error", err)
	suite.Assert().Truef(errors.Is(err, errors.Unauthorized), "Errors should be an Unauthorized Error. Error: %v", err)
	var details *RequestError
	suite.Require().True(errors.As(err, &details), "Error should be a RequestError")
}

func (suite *RequestSuite) TestShouldReceiveUnauthorizedErrorWithEmptyGrant() {
	reqURL, _ := suite.ServerURL.Parse("/details/emptygrant")
	client := NewClient(suite.Logger.ToContext(context.Background()))
	suite.Require().NotNil(client)
	_, err := client.sendRequest(context.Background(), &request.Options{
		URL:    reqURL,
		Logger: suite.Logger,
	}, nil)
	suite.Require().NotNil(err, "Should have failed sending request")
	suite.Logger.Errorf("Analyzing ", err)
	suite.Assert().Truef(errors.Is(err, errors.Unauthorized), "Errors should be an Unauthorized Error. Error: %v", err)
	suite.Assert().Truef(errors.Is(err, InvalidGrant), "Errors should be an Invalid Grant Error. Error: %v", err)
}

func (suite *RequestSuite) TestShouldReceiveNotFoundErrorWithDetails() {
	reqURL, _ := suite.ServerURL.Parse("/details/notfound")
	client := NewClient(suite.Logger.ToContext(context.Background()))
	suite.Require().NotNil(client)
	_, err := client.sendRequest(context.Background(), &request.Options{
		URL:    reqURL,
		Logger: suite.Logger,
	}, nil)
	suite.Require().NotNil(err, "Should have failed sending request")
	suite.Logger.Errorf("Analyzing Error", err)
	suite.Assert().Truef(errors.Is(err, errors.NotFound), "Errors should be a Not Found Error. Error: %v", err)
	var details *RequestError
	suite.Require().True(errors.As(err, &details), "Error should be a RequestError")
}

func (suite *RequestSuite) TestShouldReceiveErrorWithDetails() {
	reqURL, _ := suite.ServerURL.Parse("/details/forbidden")
	client := NewClient(suite.Logger.ToContext(context.Background()))
	suite.Require().NotNil(client)
	_, err := client.sendRequest(context.Background(), &request.Options{
		URL:    reqURL,
		Logger: suite.Logger,
	}, nil)
	suite.Require().NotNil(err, "Should have failed sending request")
	suite.Logger.Errorf("Analyzing Error", err)
	var details *RequestError
	suite.Require().True(errors.As(err, &details), "Error should be a RequestError")
}

func (suite *RequestSuite) TestCanMarshalRequestError() {
	payload, err := json.Marshal(InvalidGrant)
	suite.Require().Nilf(err, "Error should be nil. Error: %v", err)
	suite.Assert().NotEmpty(payload)
}

func (suite *RequestSuite) TestRequestErrorImplementsIs() {
	var err error = InvalidGrant
	suite.Assert().True(errors.Is(err, InvalidGrant))
	suite.Assert().True(errors.Is(err, &InvalidGrant))
	suite.Assert().False(errors.Is(err, FolderNotEmpty))
	suite.Assert().False(errors.Is(err, errors.NotFound))
}

// Suite Tools

func (suite *RequestSuite) SetupSuite() {
	suite.Name = strings.TrimSuffix(reflect.TypeOf(*suite).Name(), "Suite")
	suite.Logger = logger.Create("test",
		&logger.FileStream{
			Path:        fmt.Sprintf("./log/test-%s.log", strings.ToLower(suite.Name)),
			Unbuffered:  true,
			FilterLevel: logger.TRACE,
		},
	).Child("test", "test")
	suite.Logger.Infof("Suite Start: %s %s", suite.Name, strings.Repeat("=", 80-14-len(suite.Name)))
	suite.Server = suite.CreateServer()
	suite.ServerURL, _ = url.Parse(suite.Server.URL)
}

func (suite *RequestSuite) TearDownSuite() {
	if suite.T().Failed() {
		suite.Logger.Warnf("At least one test failed, we are not cleaning")
		suite.T().Log("At least one test failed, we are not cleaning")
	} else {
		suite.Logger.Infof("All tests succeeded, we are cleaning")
	}
	suite.Logger.Infof("Suite End: %s %s", suite.Name, strings.Repeat("=", 80-12-len(suite.Name)))
	suite.Logger.Close()
}

func (suite *RequestSuite) BeforeTest(suiteName, testName string) {
	suite.Logger.Infof("Test Start: %s %s", testName, strings.Repeat("-", 80-13-len(testName)))
	suite.Start = time.Now()
}

func (suite *RequestSuite) AfterTest(suiteName, testName string) {
	duration := time.Since(suite.Start)
	suite.Logger.Record("duration", duration.String()).Infof("Test End: %s %s", testName, strings.Repeat("-", 80-11-len(testName)))
}
