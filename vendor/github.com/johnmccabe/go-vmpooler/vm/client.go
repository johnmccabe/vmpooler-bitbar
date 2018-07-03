package vm

import (
	"net/http"

	"github.com/johnmccabe/go-vmpooler/client"
)

// Client is the client used for interacting with the VM API, uses a token value for auth
type Client struct {
	client.BaseClient
	Token string
}

// NewClient returns a new vm API client
func NewClient(endpoint, token string) *Client {
	baseClient := client.BaseClient{
		Client:   &http.Client{},
		Endpoint: endpoint,
	}

	c := &Client{
		BaseClient: baseClient,
		Token:      token,
	}
	return c
}

// SendWithAuth submits a request with X-AUTH-TOKEN header set
func (c *Client) SendWithAuth(req *http.Request, v interface{}) error {
	req.Header.Set("X-AUTH-TOKEN", c.Token)
	return c.Send(req, v)
}
