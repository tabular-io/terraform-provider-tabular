package tabular

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func (c *Client) GetWarehouses() ([]Warehouse, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/ws/v1/warehouses", c.Endpoint), nil)
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

	var warehouses []Warehouse
	err = json.Unmarshal(body, &warehouses)
	if err != nil {
		return nil, err
	}

	return warehouses, nil
}
