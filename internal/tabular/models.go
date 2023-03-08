package tabular

type Warehouse struct {
	Id     string
	Name   string
	Region string
}

type Role struct {
	Id       string
	Name     string
	Children []Role
	Members  []Member
}

type Member struct {
	Id        string
	Email     string
	WithGrant bool
}

type RoleRelation struct {
	ParentRoleId string
	ChildRoleId  string
}

var DatabasePrivileges = []string{
	"MODIFY_DATABASE",
	"LIST_TABLES",
	"CREATE_TABLE",
	"FUTURE_SELECT",
	"FUTURE_UPDATE",
	"FUTURE_DROP_TABLE",
}

type RoleDatabaseGrants struct {
	RoleName            string
	WarehouseId         string
	Database            string
	Privileges          []string
	PrivilegesWithGrant []string
}
