package token

import (
	"errors"
	"net/http"

	"github.com/johnmccabe/go-vmpooler/client"
)

// Client is the basic auth client used for interacting with the token API
type Client struct {
	client.BaseClient
	Username string
	Password string
}

// NewClient returns a new token API client
func NewClient(endpoint, username, password string) *Client {
	baseClient := client.BaseClient{
		Client:   &http.Client{},
		Endpoint: endpoint,
	}

	c := &Client{
		BaseClient: baseClient,
		Username:   username,
		Password:   password,
	}
	return c
}

// SendWithBasicAuth submits a request with basic auth
func (c *Client) SendWithBasicAuth(req *http.Request, v interface{}) error {
	req.SetBasicAuth(c.Username, c.Password)
	if req.URL.Scheme == "http" {
		return errors.New("basic auth not allowed on insecure transport (http)")
	}
	return c.Send(req, v)
}
