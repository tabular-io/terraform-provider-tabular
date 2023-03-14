package tabular

type (
	Warehouse struct {
		Id     string
		Name   string
		Region string
	}

	Database struct {
		WarehouseId string
		Namespace   []string
		Properties  map[string]string
	}

	CreateRole struct {
		Name string
	}

	Role struct {
		Id       string
		Name     string
		Children []Role
		Members  []Member
	}

	Member struct {
		Id        string
		Email     string
		WithAdmin bool
	}

	RoleRelation struct {
		ParentRoleId string
		ChildRoleId  string
	}

	RoleDatabaseGrants struct {
		RoleName            string
		WarehouseId         string
		Database            string
		Privileges          []string
		PrivilegesWithGrant []string
	}
)

var DatabasePrivileges = []string{
	"MODIFY_DATABASE",
	"LIST_TABLES",
	"CREATE_TABLE",
	"FUTURE_SELECT",
	"FUTURE_UPDATE",
	"FUTURE_DROP_TABLE",
}
