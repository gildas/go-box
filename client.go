package box

import (
	"net/url"
	"context"
	"github.com/gildas/go-logger"
)

// Client is the Box Client
type Client struct {
	Token  *Token         `json:"token"`
	Proxy  *url.URL       `json:"proxy"`
	Logger *logger.Logger `json:"-"`
}

// NewClient instantiates a new Client
func NewClient(ctx context.Context) (*Client) {
	log, err := logger.FromContext(ctx)
	if err != nil {
		log = logger.Create("Box")
	}
	return &Client{
		Logger: log.Topic("box").Scope("box").Child(),
	}
}