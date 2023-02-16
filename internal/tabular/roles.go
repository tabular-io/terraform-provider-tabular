package tabular

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func (c *Client) GetRole(roleId string) (*Role, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/ws/v1/scim2/Groups/%s", c.Endpoint, roleId), nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var role Role
	err = json.Unmarshal(body, &role)
	if err != nil {
		return nil, err
	}

	return &role, nil
}
