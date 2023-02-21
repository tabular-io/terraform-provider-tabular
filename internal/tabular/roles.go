package tabular

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

func (c *Client) GetRole(roleId string) (*Role, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/ws/v1/auth/roles/%s", c.Endpoint, roleId), nil)
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

	var role Role
	err = json.Unmarshal(body, &role)
	if err != nil {
		return nil, err
	}

	return &role, nil
}

type CreateRoleRequest struct {
	RoleName string `json:"roleName"`
}

func (c *Client) CreateRole(name string) (*Role, error) {
	reqBody, err := json.Marshal(CreateRoleRequest{
		RoleName: name,
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/ws/v1/auth/roles/organizations/%s", c.Endpoint, c.OrgId), bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

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

func (c *Client) DeleteRole(id string) (err error) {
	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/ws/v1/auth/roles/%s", c.Endpoint, id), nil)
	if err != nil {
		return err
	}

	// TODO: Require flag to delete role with usage?
	_, err = c.doRequest(req)
	if err != nil {
		return err
	}

	return
}
