package box

import (
	"net/url"
	"context"
	"github.com/gildas/go-logger"
)

// Client is the Box Client
type Client struct {
	Proxy       *url.URL       `json:"proxy"`
	Auth        *Auth          `json:"-"`
	Files       *Files         `json:"-"`
	SharedLinks *SharedLinks   `json:"-"`
	Logger      *logger.Logger `json:"-"`
}

// NewClient instantiates a new Client
func NewClient(ctx context.Context) (*Client) {
	log, err := logger.FromContext(ctx)
	if err != nil {
		log = logger.Create("Box")
	}
	client := &Client{
		Logger: log.Topic("box").Scope("box").Child(),
	}
	client.Auth        = &Auth{client, TokenFromContext(ctx)}
	client.Files       = &Files{client}
	client.SharedLinks = &SharedLinks{client}
	return client
}

// IsAuthenticated tells if the client is authenticated
func (client *Client) IsAuthenticated() bool {
	return client.Auth.IsAuthenticated()
}
