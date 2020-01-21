package box

import (
	"net/url"
	"context"
	"github.com/gildas/go-logger"
)

// Client is the Box Client
type Client struct {
	Api         *url.URL       `json:"api"`
	Proxy       *url.URL       `json:"proxy"`
	Auth        *Auth          `json:"-"`
	Files       *Files         `json:"-"`
	Folders     *Folders       `json:"-"`
	SharedLinks *SharedLinks   `json:"-"`
	Logger      *logger.Logger `json:"-"`
}

// NewClient instantiates a new Client
func NewClient(ctx context.Context) (*Client) {
	client := &Client{}
	log, err := logger.FromContext(ctx)
	if err != nil {
		log = logger.Create("Box")
	}
	client.Logger      = log.Child("box", "box")
	client.Api         = &url.URL{Scheme: "https", Host: "api.box.com", Path: "/2.0"}
	client.Auth        = &Auth{client, TokenFromContext(ctx)}
	client.Files       = &Files{client, client.moduleApi("files")}
	client.Folders     = &Folders{client, client.moduleApi("folders")}
	client.SharedLinks = &SharedLinks{client, client.moduleApi("files")}
	return client
}

// IsAuthenticated tells if the client is authenticated
func (client *Client) IsAuthenticated() bool {
	return client.Auth.IsAuthenticated()
}

// moduleApi computes the API URL of the given module
func (client *Client) moduleApi(name string) *url.URL {
	api, _ := client.Api.Parse(name)
	return api
}