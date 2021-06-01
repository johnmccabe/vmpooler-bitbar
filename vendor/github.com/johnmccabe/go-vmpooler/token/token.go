package token

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	poolertime "github.com/johnmccabe/go-vmpooler/time"
)

// Token represents a Token known to vmpooler
// If Last is nil then the token has never been used
type Token struct {
	Token   string
	User    string
	Created time.Time
	Last    *time.Time
}

type createTokenResponse struct {
	Ok    bool
	Token string
}

type deleteTokenResponse struct {
	Ok bool
}

type legacyGetAllTokenOutput struct {
	User    string
	Created poolertime.PoolerTime
	Last    poolertime.PoolerTime
}

// GenerateToken creates a new token for the authenticated user
// Endpoint: POST /v1/token
func (c *Client) Generate() (*Token, error) {
	req, err := c.NewRequest("POST", fmt.Sprintf("%s%s", c.Endpoint, "/token"), nil)
	if err != nil {
		return nil, err
	}

	response := createTokenResponse{}

	err = c.SendWithBasicAuth(req, &response)
	if err != nil {
		return nil, err
	}

	if !response.Ok {
		return nil, errors.New("unable to parse create token response")
	}

	return c.Get(response.Token)
}

// Delete the specified token
// Endpoint: DELETE /v1/token/<token>
// TODO - require a force flag to delete a token which has VMs allocated?
func (c *Client) Delete(token string) error {
	req, err := c.NewRequest("DELETE", fmt.Sprintf("%s%s/%s", c.Endpoint, "/token", token), nil)
	if err != nil {
		return err
	}

	response := deleteTokenResponse{}

	err = c.SendWithBasicAuth(req, &response)
	if err != nil {
		return err
	}

	if !response.Ok {
		return errors.New("unable to delete token")
	}

	return nil
}

// Get returns details of the specified token value
// Endpoint: GET /v1/token/<token>
func (c *Client) Get(token string) (*Token, error) {
	req, err := c.NewRequest("GET", fmt.Sprintf("%s%s/%s", c.Endpoint, "/token", token), nil)
	if err != nil {
		return nil, err
	}

	response := map[string]interface{}{}

	err = c.SendWithBasicAuth(req, &response)
	if err != nil {
		return nil, err
	}

	if err := responseOk(response); err != nil {
		return nil, err
	}

	lgto := legacyGetAllTokenOutput{}
	if l, ok := response[token]; !ok {
		return nil, errors.New("unable to parse response")
	} else {
		b, err := json.Marshal(l)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(b, &lgto); err != nil {
			return nil, err
		}
	}

	t := &Token{
		Token:   token,
		User:    lgto.User,
		Created: lgto.Created.Time,
	}

	if !lgto.Last.IsZero() {
		t.Last = &lgto.Last.Time
	}
	return t, nil
}

// GetAll returns details of all tokens belonging to the authenticated user
// Endpoint: GET /v1/token/
func (c *Client) GetAll() ([]Token, error) {
	req, err := c.NewRequest("GET", fmt.Sprintf("%s%s", c.Endpoint, "/token"), nil)
	if err != nil {
		return nil, err
	}

	response := map[string]interface{}{}

	err = c.SendWithBasicAuth(req, &response)
	if err != nil {
		return nil, err
	}

	if err := responseOk(response); err != nil {
		return nil, err
	}

	tokenIds := []string{}
	for k := range response {
		if k == "ok" {
			continue
		}
		tokenIds = append(tokenIds, k)
	}

	tokens := []Token{}
	for _, tokenId := range tokenIds {
		token, err := c.Get(tokenId)
		if err != nil {
			return nil, err
		}
		tokens = append(tokens, *token)
	}

	return tokens, nil
}

func responseOk(response map[string]interface{}) error {
	if r, ok := response["ok"]; ok {
		if rb, ok := r.(bool); !ok || !rb {
			return errors.New("response from vmpooler not OK")
		}
	} else {
		return errors.New("invalid response returned from vmpooler")
	}
	return nil
}
