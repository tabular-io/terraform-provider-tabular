package tabular

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
)

type roleRef struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type roleDatabaseGrantDetail struct {
	Role      roleRef `json:"role"`
	Privilege string  `json:"privilege"`
	WithGrant bool    `json:"withGrant"`
}

type changeRoleGrantRequest struct {
	Role      string `json:"roleName"`
	Privilege string `json:"privilege"`
	WithGrant bool   `json:"withGrant"`
}

func (c *Client) GetRoleDatabaseGrants(warehouseId, database, roleName string) (*RoleDatabaseGrants, error) {
	req, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("%s/ws/v1/grants/warehouses/%s/namespaces/%s/grants", c.Endpoint, warehouseId, database),
		nil,
	)
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

	var grantsResp []roleDatabaseGrantDetail
	err = json.Unmarshal(body, &grantsResp)
	if err != nil {
		return nil, err
	}

	grants := RoleDatabaseGrants{
		WarehouseId:         warehouseId,
		Database:            database,
		RoleName:            roleName,
		Privileges:          []string{},
		PrivilegesWithGrant: []string{},
	}
	for _, g := range grantsResp {
		if g.Role.Name == roleName {
			if g.WithGrant {
				grants.PrivilegesWithGrant = append(grants.PrivilegesWithGrant, g.Privilege)
			} else {
				grants.Privileges = append(grants.Privileges, g.Privilege)
			}
		}
	}
	sort.Strings(grants.Privileges)
	sort.Strings(grants.PrivilegesWithGrant)

	return &grants, nil
}

func (c *Client) AddRoleDatabaseGrants(warehouseId, database, roleName string, privileges []string, withGrant bool) (err error) {
	if privileges == nil || len(privileges) == 0 {
		return
	}
	var grants []changeRoleGrantRequest
	for _, priv := range privileges {
		grants = append(grants, changeRoleGrantRequest{
			Role:      roleName,
			Privilege: priv,
			WithGrant: withGrant,
		})
	}

	reqBody, err := json.Marshal(&grants)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(
		http.MethodPut,
		fmt.Sprintf("%s/ws/v1/grants/warehouses/%s/namespaces/%s/grants", c.Endpoint, warehouseId, database),
		bytes.NewReader(reqBody),
	)
	if err != nil {
		return err
	}

	_, err = c.doRequest(req)
	if err != nil {
		clientErr, ok := err.(*ClientError)
		if ok && clientErr.response.StatusCode == 404 {
			return nil
		} else {
			return err
		}
	}
	return
}

func (c *Client) RevokeRoleDatabaseGrants(warehouseId, database, roleName string, privileges []string, withGrant bool) (err error) {
	if privileges == nil || len(privileges) == 0 {
		return
	}
	var grants []changeRoleGrantRequest
	for _, priv := range privileges {
		grants = append(grants, changeRoleGrantRequest{
			Role:      roleName,
			Privilege: priv,
			WithGrant: withGrant,
		})
	}

	reqBody, err := json.Marshal(&grants)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(
		http.MethodDelete,
		fmt.Sprintf("%s/ws/v1/grants/warehouses/%s/namespaces/%s/grants", c.Endpoint, warehouseId, database),
		bytes.NewReader(reqBody),
	)
	if err != nil {
		return err
	}

	_, err = c.doRequest(req)
	if err != nil {
		clientErr, ok := err.(*ClientError)
		if ok && clientErr.response.StatusCode == 404 {
			return nil
		} else {
			return err
		}
	}
	return
}
