package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// BaseClient implements the client functionality common to the token/basic auth variants
// used by the vm and token APIs
type BaseClient struct {
	Client   *http.Client
	Endpoint string
}

// NewRequest constructs a request
// Convert payload to a JSON
func (c *BaseClient) NewRequest(method, url string, payload interface{}) (*http.Request, error) {
	var buf io.Reader
	if payload != nil {
		var b []byte
		b, err := json.Marshal(&payload)
		if err != nil {
			return nil, err
		}
		buf = bytes.NewBuffer(b)
	}
	return http.NewRequest(method, url, buf)
}

// Send makes a request to the API, the response body will be
// unmarshaled into v, or if v is an io.Writer, the response will
// be written to it without decoding
func (c *BaseClient) Send(req *http.Request, v interface{}) error {
	var (
		err  error
		resp *http.Response
	)

	resp, err = c.Client.Do(req)

	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("%d", resp.StatusCode)
	}

	if v != nil {
		if w, ok := v.(io.Writer); ok {
			io.Copy(w, resp.Body)
		} else {
			err = json.NewDecoder(resp.Body).Decode(v)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
