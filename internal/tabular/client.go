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
	HTTPClient *http.Client
}

func NewClient(endpoint, tokenEndpoint, credential string) (*Client, error) {
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
		HTTPClient: clientConfig.Client(context.Background()),
	}
	return &client, nil
}

type ClientError struct {
	statusCode   int
	responseBody string
	response     http.Response
}

func (err *ClientError) Error() string {
	return fmt.Sprintf("[%d] %s", err.statusCode, err.responseBody)
}

func (c *Client) doRequest(req *http.Request) ([]byte, error) {
	if req.Body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, &ClientError{
			statusCode:   resp.StatusCode,
			response:     *resp,
			responseBody: string(body[:]),
		}
	}

	return body, err
}
