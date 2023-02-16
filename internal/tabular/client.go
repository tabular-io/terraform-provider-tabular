package tabular

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type Client struct {
	Endpoint   string
	username   string
	password   string
	HTTPClient *http.Client
}

func NewClient(endpoint, credential string) (*Client, error) {
	parts := strings.SplitN(credential, ":", 2)
	if len(parts) != 2 {
		return nil, errors.New("bad credential provided")
	}
	client := Client{
		Endpoint:   endpoint,
		username:   parts[0],
		password:   parts[1],
		HTTPClient: http.DefaultClient,
	}
	return &client, nil
}

func (c *Client) doRequest(req *http.Request) ([]byte, error) {
	req.SetBasicAuth(c.username, c.password)
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status: %d, body: %s", resp.StatusCode, body)
	}

	return body, err
}
