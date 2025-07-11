package base_menu_btn

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gbrayhan/microservices-go/src/domain"

	domainErrors "github.com/gbrayhan/microservices-go/src/domain/errors"
	domainMenu "github.com/gbrayhan/microservices-go/src/domain/sys/menu"
	logger "github.com/gbrayhan/microservices-go/src/infrastructure/logger"
	"github.com/gbrayhan/microservices-go/src/infrastructure/repository/utils"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type SysBaseMenuBtn struct {
	ID               int            `gorm:"column:id;primary_key" json:"id"`
	CreatedAt        time.Time      `gorm:"column:created_at" json:"createdAt,omitempty"`
	UpdatedAt        time.Time      `gorm:"column:updated_at" json:"updatedAt,omitempty"`
	DeletedAt        gorm.DeletedAt `gorm:"column:deleted_at;index:idx_sys_apis_deleted_at" json:"deletedAt,omitempty"`
	Name             string         `gorm:"column:name" json:"name,omitempty"`
	Desc             string         `gorm:"column:desc" json:"desc,omitempty"`
	SysBaseMenuBtnID int64          `gorm:"column:sys_base_menu_id" json:"sysBaseMenuId,omitempty"`
}

func (SysBaseMenuBtn) TableName() string {
	return "sys_menu_btns"
}

var ColumnsMenuMapping = map[string]string{
	"id":          "id",
	"path":        "path",
	"menuName":    "menu_name",
	"description": "description",
	"menuGroup":   "menu_group",
	"method":      "method",
	"createdAt":   "created_at",
	"updatedAt":   "updated_at",
}

// MenuRepositoryInterface defines the interface for menu repository operations
type MenuRepositoryInterface interface {
	GetAll() (*[]domainMenu.Menu, error)
	Create(menuDomain *domainMenu.Menu) (*domainMenu.Menu, error)
	GetByID(id int) (*domainMenu.Menu, error)
	Update(id int, menuMap map[string]interface{}) (*domainMenu.Menu, error)
	Delete(id int) error
	SearchPaginated(filters domain.DataFilters) (*domain.PaginatedResult[domainMenu.Menu], error)
	SearchByProperty(property string, searchText string) (*[]string, error)
	GetOneByMap(menuMap map[string]interface{}) (*domainMenu.Menu, error)
}

type Repository struct {
	DB     *gorm.DB
	Logger *logger.Logger
}

func NewMenuRepository(db *gorm.DB, loggerInstance *logger.Logger) MenuRepositoryInterface {
	return &Repository{DB: db, Logger: loggerInstance}
}

func (r *Repository) GetAll() (*[]domainMenu.Menu, error) {
	var menus []SysBaseMenuBtn
	if err := r.DB.Find(&menus).Error; err != nil {
		r.Logger.Error("Error getting all menus", zap.Error(err))
		return nil, domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}
	r.Logger.Info("Successfully retrieved all menus", zap.Int("count", len(menus)))
	return arrayToDomainMapper(&menus), nil
}

func (r *Repository) Create(menuDomain *domainMenu.Menu) (*domainMenu.Menu, error) {
	r.Logger.Info("Creating new menu", zap.String("path", menuDomain.Path))
	menuRepository := fromDomainMapper(menuDomain)
	txDb := r.DB.Create(menuRepository)
	err := txDb.Error
	if err != nil {
		r.Logger.Error("Error creating menu", zap.Error(err), zap.String("Path", menuDomain.Path))
		byteErr, _ := json.Marshal(err)
		var newError domainErrors.GormErr
		errUnmarshal := json.Unmarshal(byteErr, &newError)
		if errUnmarshal != nil {
			return &domainMenu.Menu{}, errUnmarshal
		}
		switch newError.Number {
		case 1062:
			err = domainErrors.NewAppErrorWithType(domainErrors.ResourceAlreadyExists)
			return &domainMenu.Menu{}, err
		default:
			err = domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
		}
	}
	r.Logger.Info("Successfully created menu", zap.String("Path", menuDomain.Path), zap.Int("id", int(menuRepository.ID)))
	return menuRepository.toDomainMapper(), err
}

func (r *Repository) GetByID(id int) (*domainMenu.Menu, error) {
	var menu SysBaseMenuBtn
	err := r.DB.Where("id = ?", id).First(&menu).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			r.Logger.Warn("Menu not found", zap.Int("id", id))
			err = domainErrors.NewAppErrorWithType(domainErrors.NotFound)
		} else {
			r.Logger.Error("Error getting menu by ID", zap.Error(err), zap.Int("id", id))
			err = domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
		}
		return &domainMenu.Menu{}, err
	}
	r.Logger.Info("Successfully retrieved menu by ID", zap.Int("id", id))
	return menu.toDomainMapper(), nil
}

func (r *Repository) Update(id int, menuMap map[string]interface{}) (*domainMenu.Menu, error) {
	var menuObj SysBaseMenuBtn
	menuObj.ID = id
	err := r.DB.Model(&menuObj).Updates(menuMap).Error
	if err != nil {
		r.Logger.Error("Error updating menu", zap.Error(err), zap.Int("id", id))
		byteErr, _ := json.Marshal(err)
		var newError domainErrors.GormErr
		errUnmarshal := json.Unmarshal(byteErr, &newError)
		if errUnmarshal != nil {
			return &domainMenu.Menu{}, errUnmarshal
		}
		switch newError.Number {
		case 1062:
			return &domainMenu.Menu{}, domainErrors.NewAppErrorWithType(domainErrors.ResourceAlreadyExists)
		default:
			return &domainMenu.Menu{}, domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
		}
	}
	if err := r.DB.Where("id = ?", id).First(&menuObj).Error; err != nil {
		r.Logger.Error("Error retrieving updated menu", zap.Error(err), zap.Int("id", id))
		return &domainMenu.Menu{}, err
	}
	r.Logger.Info("Successfully updated menu", zap.Int("id", id))
	return menuObj.toDomainMapper(), nil
}

func (r *Repository) Delete(id int) error {
	tx := r.DB.Delete(&SysBaseMenuBtn{}, id)
	if tx.Error != nil {
		r.Logger.Error("Error deleting menu", zap.Error(tx.Error), zap.Int("id", id))
		return domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}
	if tx.RowsAffected == 0 {
		r.Logger.Warn("Menu not found for deletion", zap.Int("id", id))
		return domainErrors.NewAppErrorWithType(domainErrors.NotFound)
	}
	r.Logger.Info("Successfully deleted menu", zap.Int("id", id))
	return nil
}

func (r *Repository) SearchPaginated(filters domain.DataFilters) (*domain.PaginatedResult[domainMenu.Menu], error) {
	query := r.DB.Model(&SysBaseMenuBtn{})

	// Apply like filters
	for field, values := range filters.LikeFilters {
		if len(values) > 0 {
			for _, value := range values {
				if value != "" {
					column := ColumnsMenuMapping[field]
					if column != "" {
						query = query.Where(column+" ILIKE ?", "%"+value+"%")
					}
				}
			}
		}
	}

	// Apply exact matches
	for field, values := range filters.Matches {
		if len(values) > 0 {
			column := ColumnsMenuMapping[field]
			if column != "" {
				query = query.Where(column+" IN ?", values)
			}
		}
	}

	// Apply date range filters
	for _, dateFilter := range filters.DateRangeFilters {
		column := ColumnsMenuMapping[dateFilter.Field]
		if column != "" {
			if dateFilter.Start != nil {
				query = query.Where(column+" >= ?", dateFilter.Start)
			}
			if dateFilter.End != nil {
				query = query.Where(column+" <= ?", dateFilter.End)
			}
		}
	}

	// Apply sorting
	if len(filters.SortBy) > 0 && filters.SortDirection.IsValid() {
		for _, sortField := range filters.SortBy {
			column := ColumnsMenuMapping[sortField]
			if column != "" {
				query = query.Order(column + " " + string(filters.SortDirection))
			}
		}
	}

	// Count total records
	var total int64
	clonedQuery := query
	clonedQuery.Count(&total)

	// Apply pagination
	if filters.Page < 1 {
		filters.Page = 1
	}
	if filters.PageSize < 1 {
		filters.PageSize = 10
	}
	offset := (filters.Page - 1) * filters.PageSize

	var menus []SysBaseMenuBtn
	if err := query.Offset(offset).Limit(filters.PageSize).Find(&menus).Error; err != nil {
		r.Logger.Error("Error searching menus", zap.Error(err))
		return nil, domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}

	totalPages := int((total + int64(filters.PageSize) - 1) / int64(filters.PageSize))

	result := &domain.PaginatedResult[domainMenu.Menu]{
		Data:       arrayToDomainMapper(&menus),
		Total:      total,
		Page:       filters.Page,
		PageSize:   filters.PageSize,
		TotalPages: totalPages,
	}

	r.Logger.Info("Successfully searched menus",
		zap.Int64("total", total),
		zap.Int("page", filters.Page),
		zap.Int("pageSize", filters.PageSize))

	return result, nil
}

func (r *Repository) SearchByProperty(property string, searchText string) (*[]string, error) {
	column := ColumnsMenuMapping[property]
	if column == "" {
		r.Logger.Warn("Invalid property for search", zap.String("property", property))
		return nil, domainErrors.NewAppErrorWithType(domainErrors.ValidationError)
	}

	var coincidences []string
	if err := r.DB.Model(&SysBaseMenuBtn{}).
		Distinct(column).
		Where(column+" ILIKE ?", "%"+searchText+"%").
		Limit(20).
		Pluck(column, &coincidences).Error; err != nil {
		r.Logger.Error("Error searching by property", zap.Error(err), zap.String("property", property))
		return nil, domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}

	r.Logger.Info("Successfully searched by property",
		zap.String("property", property),
		zap.Int("results", len(coincidences)))

	return &coincidences, nil
}

func (u *SysBaseMenuBtn) toDomainMapper() *domainMenu.Menu {
	return &domainMenu.Menu{
		ID:        u.ID,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

func fromDomainMapper(u *domainMenu.Menu) *SysBaseMenuBtn {
	return &SysBaseMenuBtn{
		ID:        u.ID,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

func arrayToDomainMapper(menus *[]SysBaseMenuBtn) *[]domainMenu.Menu {
	menusDomain := make([]domainMenu.Menu, len(*menus))
	for i, menu := range *menus {
		menusDomain[i] = *menu.toDomainMapper()
	}
	return &menusDomain
}

func (r *Repository) GetOneByMap(menuMap map[string]interface{}) (*domainMenu.Menu, error) {
	var menuRepository SysBaseMenuBtn
	tx := r.DB.Limit(1)
	for key, value := range menuMap {
		if !utils.IsZeroValue(value) {
			tx = tx.Where(fmt.Sprintf("%s = ?", key), value)
		}
	}
	if err := tx.Find(&menuRepository).Error; err != nil {
		return &domainMenu.Menu{}, domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}
	return menuRepository.toDomainMapper(), nil
}
