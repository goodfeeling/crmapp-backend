package role

import (
	"github.com/gbrayhan/microservices-go/src/domain"
)

type Role struct {
	ID            int64
	Name          string
	ParentID      int64
	DefaultRouter string
	Status        bool
	Order         int64
	Label         string
	Description   string
	CreatedAt     domain.CustomTime
	UpdatedAt     domain.CustomTime
}
type SearchResultRole struct {
	Data       *[]Role `json:"data"`
	Total      int64   `json:"total"`
	Page       int     `json:"page"`
	PageSize   int     `json:"page_size"`
	TotalPages int     `json:"total_page"`
}

type RoleNode struct {
	ID       string      `json:"value"`
	Name     string      `json:"title"`
	Key      string      `json:"key"`
	Path     []int64     `json:"path"`
	Children []*RoleNode `json:"children"`
}

type RoleTree struct {
	ID            int64             `json:"id"`
	Name          string            `json:"name"`
	ParentID      int64             `json:"parent_id"`
	DefaultRouter string            `json:"default_router"`
	Status        bool              `json:"status"`
	Order         int64             `json:"order"`
	Label         string            `json:"label"`
	Description   string            `json:"description"`
	CreatedAt     domain.CustomTime `json:"created_at"`
	UpdatedAt     domain.CustomTime `json:"updated_at"`
	Path          []int64           `json:"path"`
	Children      []*RoleTree       `json:"children"`
}

type IRoleService interface {
	GetAll() ([]*RoleTree, error)
	GetByID(id int) (*Role, error)
	Create(newRole *Role) (*Role, error)
	Delete(id int) error
	Update(id int, userMap map[string]interface{}) (*Role, error)
	SearchPaginated(filters domain.DataFilters) (*SearchResultRole, error)
	SearchByProperty(property string, searchText string) (*[]string, error)
	GetOneByMap(userMap map[string]interface{}) (*Role, error)
	GetTreeRoles() (*RoleNode, error)
}
