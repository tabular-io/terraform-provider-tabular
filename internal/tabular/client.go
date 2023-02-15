package tabular

import (
	"context"
	"errors"
	"fmt"
	"golang.org/x/oauth2/clientcredentials"
	"io"
	"net/http"
	"strings"
)

type Client struct {
	Endpoint   string
	OrgId      string
	HTTPClient *http.Client
}

func NewClient(endpoint, tokenEndpoint, orgId, credential string) (*Client, error) {
	parts := strings.SplitN(credential, ":", 2)
	if len(parts) != 2 {
		return nil, errors.New("bad credential provided")
	}

	clientConfig := clientcredentials.Config{
		ClientID:     parts[0],
		ClientSecret: parts[1],
		TokenURL:     tokenEndpoint,
	}

	client := Client{
		Endpoint:   endpoint,
		OrgId:      orgId,
		HTTPClient: clientConfig.Client(context.Background()),
	}
	return &client, nil
}

type ClientError struct {
	message  string
	response http.Response
}

func (err *ClientError) Error() string {
	return ""
}

func (c *Client) doRequest(req *http.Request) ([]byte, error) {
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
		return nil, &ClientError{
			message:  fmt.Sprintf("Error Code: %d", resp.StatusCode),
			response: *resp,
		}
	}

	return body, err
}
