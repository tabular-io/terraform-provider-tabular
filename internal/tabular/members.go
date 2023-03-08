package tabular

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

func (c *Client) GetOrgMemberIdsMap() (map[string]string, error) {
	orgMembers, err := c.getOrgMembers()
	if err != nil {
		return nil, err
	}
	orgMemberMap := make(map[string]string, len(orgMembers))
	for _, orgMember := range orgMembers {
		orgMemberMap[orgMember.Email] = orgMember.Id
	}
	return orgMemberMap, nil
}

func (c *Client) getOrgMembers() ([]Member, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/ws/v1/grants/members", c.Endpoint), nil)
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

	var members []Member
	err = json.Unmarshal(body, &members)
	if err != nil {
		return nil, err
	}

	return members, nil
}

func (c *Client) AddRoleMembers(roleName string, adminMemberIds, memberIds []string) (err error) {
	if (adminMemberIds == nil || len(adminMemberIds) == 0) && (memberIds == nil || len(memberIds) == 0) {
		return
	}
	var request []AddRoleMemberRequest
	for _, member := range adminMemberIds {
		request = append(request, AddRoleMemberRequest{
			MemberId: member,
			IsAdmin:  true,
		})
	}
	for _, member := range memberIds {
		request = append(request, AddRoleMemberRequest{
			MemberId: member,
			IsAdmin:  false,
		})
	}
	reqBody, err := json.Marshal(request)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/ws/v1/grants/roles/%s/members", c.Endpoint, roleName), bytes.NewReader(reqBody))
	if err != nil {
		return err
	}

	_, err = c.doRequest(req)
	if err != nil {
		return err
	}

	return
}

func (c *Client) DeleteRoleMembers(roleName string, memberIds []string) (err error) {
	if memberIds == nil || len(memberIds) == 0 {
		return
	}
	reqBody, err := json.Marshal(memberIds)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/ws/v1/grants/roles/%s/members", c.Endpoint, roleName), bytes.NewReader(reqBody))
	if err != nil {
		return err
	}

	_, err = c.doRequest(req)
	if err != nil {
		return err
	}

	return
}
