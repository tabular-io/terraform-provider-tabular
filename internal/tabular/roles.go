package tabular

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

type CreateRoleRequest struct {
	RoleName string `json:"roleName"`
}

type UpdateRoleRequest struct {
	RoleName string `json:"roleName"`
}

type AddRoleMemberRequest struct {
	MemberId string `json:"memberId"`
	IsAdmin  bool   `json:"withAdmin"`
}

func (c *Client) GetRole(roleName string) (*Role, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/ws/v1/grants/roles/%s", c.Endpoint, roleName), nil)
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

func (c *Client) CreateRole(name string) (*Role, error) {
	reqBody, err := json.Marshal(CreateRoleRequest{
		RoleName: name,
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/ws/v1/grants/roles", c.Endpoint), bytes.NewReader(reqBody))
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

func (c *Client) RenameRole(roleName string, newRoleName string) (*Role, error) {
	reqBody, err := json.Marshal(UpdateRoleRequest{
		RoleName: newRoleName,
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/ws/v1/grants/roles/%s", c.Endpoint, roleName), bytes.NewReader(reqBody))
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

func (c *Client) DeleteRole(roleName string, force bool) (err error) {
	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/ws/v1/grants/roles/%s", c.Endpoint, roleName), nil)
	if err != nil {
		return err
	}
	query := req.URL.Query()
	query.Add("force", strconv.FormatBool(force))
	req.URL.RawQuery = query.Encode()

	_, err = c.doRequest(req)
	if err != nil {
		return err
	}

	return
}

func (c *Client) AddRoleRelation(parentRoleName, childRoleName string) (err error) {
	reqBody, err := json.Marshal(UpdateRoleRequest{
		RoleName: childRoleName,
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/ws/v1/grants/roles/%s/children", c.Endpoint, parentRoleName), bytes.NewReader(reqBody))
	if err != nil {
		return err
	}

	_, err = c.doRequest(req)
	if err != nil {
		return err
	}

	return
}

func (c *Client) DeleteRoleRelation(parentRoleName, childRoleName string) (err error) {
	reqBody, err := json.Marshal(UpdateRoleRequest{
		RoleName: childRoleName,
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/ws/v1/grants/roles/%s/children", c.Endpoint, parentRoleName), bytes.NewReader(reqBody))
	if err != nil {
		return err
	}

	_, err = c.doRequest(req)
	if err != nil {
		return err
	}

	return
}
