package tabular

type CreateRole struct {
	Name string
}

type Role struct {
	Id       string
	Name     string
	Children *[]Role
}

type RoleRelation struct {
	ParentRoleId string
	ChildRoleId  string
}
