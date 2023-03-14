package tabular

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type databaseRequest struct {
	Namespace []string `json:"namespace"`
}

func (c *Client) GetDatabase(warehouseId, namespace string) (*Database, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/ws/v1/warehouses/%s/namespaces/%s/ext", c.Endpoint, warehouseId, namespace), nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		clientErr, ok := err.(*ClientError)
		if ok && clientErr.response.StatusCode == 404 {
			return nil, nil
		} else {
			return nil, err
		}
	}

	var database Database
	err = json.Unmarshal(body, &database)
	if err != nil {
		return nil, err
	}

	return &database, nil
}

func (c *Client) CreateDatabase(warehouseId, namespace string) (*Database, error) {
	reqBody, err := json.Marshal(databaseRequest{
		Namespace: []string{namespace},
	})
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("%s/ws/v1/warehouses/%s/namespaces/ext", c.Endpoint, warehouseId),
		bytes.NewReader(reqBody),
	)
	if err != nil {
		return nil, err
	}

	// The create database extension doesn't return the location property, so we need to refetch for it
	_, err = c.doRequest(req)
	if err != nil {
		clientErr, ok := err.(*ClientError)
		if ok && clientErr.response.StatusCode == 404 {
			return nil, nil
		} else {
			return nil, err
		}
	}

	return c.GetDatabase(warehouseId, namespace)
}

func (c *Client) DeleteDatabase(warehouseId, namespace string) (err error) {
	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/ws/v1/ice/warehouses/%s/namespaces/%s", c.Endpoint, warehouseId, namespace), nil)
	if err != nil {
		return err
	}

	_, err = c.doRequest(req)
	if err != nil {
		return err
	}

	return
}
