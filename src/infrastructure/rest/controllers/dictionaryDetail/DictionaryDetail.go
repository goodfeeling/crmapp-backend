package dictionary_detail

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gbrayhan/microservices-go/src/domain"
	domainErrors "github.com/gbrayhan/microservices-go/src/domain/errors"
	domainDictionary "github.com/gbrayhan/microservices-go/src/domain/sys/dictionary_detail"
	logger "github.com/gbrayhan/microservices-go/src/infrastructure/logger"
	dictionaryRepo "github.com/gbrayhan/microservices-go/src/infrastructure/repository/psql/sys/dictionary_detail"
	"github.com/gbrayhan/microservices-go/src/infrastructure/rest/controllers"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Structures
type DeleteBatchDictionaryDetailRequest struct {
	IDS []int `json:"ids"`
}

// Structures
type NewDictionaryRequest struct {
	ID              int    `json:"id"`
	Label           string `json:"label"  binding:"required"`
	Value           string `json:"value"  binding:"required"`
	Extend          string `json:"extend"  binding:"required"`
	Status          int16  `json:"status"  binding:"required"`
	Sort            int8   `json:"sort"  binding:"required"`
	SysDictionaryID int64  `json:"sys_dictionary_id"  binding:"required"`
}

type ResponseDictionary struct {
	ID              int               `json:"id"`
	Label           string            `json:"label"`
	Value           string            `json:"value"`
	Extend          string            `json:"extend"`
	Status          int16             `json:"status"`
	Sort            int8              `json:"sort"`
	SysDictionaryID int64             `json:"sys_dictionary_id"`
	CreatedAt       domain.CustomTime `json:"created_at,omitempty"`
	UpdatedAt       domain.CustomTime `json:"updated_at,omitempty"`
}
type IIDictionaryDetailController interface {
	NewDictionary(ctx *gin.Context)
	GetAllDictionaries(ctx *gin.Context)
	GetDictionariesByID(ctx *gin.Context)
	UpdateDictionary(ctx *gin.Context)
	DeleteDictionary(ctx *gin.Context)
	SearchPaginated(ctx *gin.Context)
	SearchByProperty(ctx *gin.Context)
	DeleteDictionaryDetails(ctx *gin.Context)
}
type DictionaryDetailController struct {
	dictionaryService domainDictionary.IDictionaryService
	Logger            *logger.Logger
}

func NewIDictionaryDetailController(dictionaryService domainDictionary.IDictionaryService, loggerInstance *logger.Logger) IIDictionaryDetailController {
	return &DictionaryDetailController{dictionaryService: dictionaryService, Logger: loggerInstance}
}

// CreateDictionary
// @Summary create dictionary
// @Description create dictionary
// @Tags dictionary create
// @Accept json
// @Produce json
// @Param book body NewDictionaryRequest true  "JSON Data"
// @Success 200 {object} controllers.CommonResponseBuilder
// @Router /v1/dictionary [post]
func (c *DictionaryDetailController) NewDictionary(ctx *gin.Context) {
	c.Logger.Info("Creating new dictionary")
	var request NewDictionaryRequest
	if err := controllers.BindJSON(ctx, &request); err != nil {
		c.Logger.Error("Error binding JSON for new dictionary", zap.Error(err))
		appError := domainErrors.NewAppError(err, domainErrors.ValidationError)
		_ = ctx.Error(appError)
		return
	}
	dictionaryModel, err := c.dictionaryService.Create(toUsecaseMapper(&request))
	if err != nil {
		c.Logger.Error("Error creating dictionary", zap.Error(err), zap.String("Label", request.Label))
		_ = ctx.Error(err)
		return
	}
	dictionaryResponse := controllers.NewCommonResponseBuilder[*ResponseDictionary]().
		Data(domainToResponseMapper(dictionaryModel)).
		Message("success").
		Status(0).
		Build()
	c.Logger.Info("Dictionary created successfully", zap.String("Label", request.Label), zap.Int("id", int(dictionaryModel.ID)))
	ctx.JSON(http.StatusOK, dictionaryResponse)
}

// GetAllDictionaries
// @Summary get all dictionaries by
// @Description get  all dictionaries by where
// @Tags dictionaries
// @Accept json
// @Produce json
// @Success 200 {object} domain.CommonResponse[[]domainDictionary.Dictionary]
// @Router /v1/dictionary [get]
func (c *DictionaryDetailController) GetAllDictionaries(ctx *gin.Context) {
	c.Logger.Info("Getting all dictionaries")
	dictionaries, err := c.dictionaryService.GetAll()
	if err != nil {
		c.Logger.Error("Error getting all dictionaries", zap.Error(err))
		appError := domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
		_ = ctx.Error(appError)
		return
	}
	c.Logger.Info("Successfully retrieved all dictionaries", zap.Int("count", len(*dictionaries)))
	ctx.JSON(http.StatusOK, domain.CommonResponse[*[]domainDictionary.Dictionary]{
		Data: dictionaries,
	})
}

// GetDictionarysByID
// @Summary get dictionaries
// @Description get dictionaries by id
// @Tags dictionaries
// @Accept json
// @Produce json
// @Success 200 {object} ResponseDictionary
// @Router /v1/dictionary/{id} [get]
func (c *DictionaryDetailController) GetDictionariesByID(ctx *gin.Context) {
	dictionaryID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		c.Logger.Error("Invalid dictionary ID parameter", zap.Error(err), zap.String("id", ctx.Param("id")))
		appError := domainErrors.NewAppError(errors.New("dictionary id is invalid"), domainErrors.ValidationError)
		_ = ctx.Error(appError)
		return
	}
	c.Logger.Info("Getting dictionary by ID", zap.Int("id", dictionaryID))
	dictionary, err := c.dictionaryService.GetByID(dictionaryID)
	if err != nil {
		c.Logger.Error("Error getting dictionary by ID", zap.Error(err), zap.Int("id", dictionaryID))
		_ = ctx.Error(err)
		return
	}
	c.Logger.Info("Successfully retrieved dictionary by ID", zap.Int("id", dictionaryID))
	ctx.JSON(http.StatusOK, domainToResponseMapper(dictionary))
}

// UpdateDictionary
// @Summary update dictionary
// @Description update dictionary
// @Tags dictionary
// @Accept json
// @Produce json
// @Param book body map[string]any  true  "JSON Data"
// @Success 200 {array} controllers.CommonResponseBuilder[ResponseDictionary]
// @Router /v1/dictionary [put]
func (c *DictionaryDetailController) UpdateDictionary(ctx *gin.Context) {
	dictionaryID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		c.Logger.Error("Invalid dictionary ID parameter for update", zap.Error(err), zap.String("id", ctx.Param("id")))
		appError := domainErrors.NewAppError(errors.New("param id is necessary"), domainErrors.ValidationError)
		_ = ctx.Error(appError)
		return
	}
	c.Logger.Info("Updating dictionary", zap.Int("id", dictionaryID))
	var requestMap map[string]any
	err = controllers.BindJSONMap(ctx, &requestMap)
	if err != nil {
		c.Logger.Error("Error binding JSON for dictionary update", zap.Error(err), zap.Int("id", dictionaryID))
		appError := domainErrors.NewAppError(err, domainErrors.ValidationError)
		_ = ctx.Error(appError)
		return
	}
	err = updateValidation(requestMap)
	if err != nil {
		c.Logger.Error("Validation error for dictionary update", zap.Error(err), zap.Int("id", dictionaryID))
		_ = ctx.Error(err)
		return
	}
	dictionaryUpdated, err := c.dictionaryService.Update(dictionaryID, requestMap)
	if err != nil {
		c.Logger.Error("Error updating dictionary", zap.Error(err), zap.Int("id", dictionaryID))
		_ = ctx.Error(err)
		return
	}
	response := controllers.NewCommonResponseBuilder[*ResponseDictionary]().
		Data(domainToResponseMapper(dictionaryUpdated)).
		Message("success").
		Status(0).
		Build()
	c.Logger.Info("Dictionary updated successfully", zap.Int("id", dictionaryID))
	ctx.JSON(http.StatusOK, response)
}

// DeleteDictionary
// @Summary delete dictionary
// @Description delete dictionary by id
// @Tags dictionary
// @Accept json
// @Produce json
// @Success 200 {object} domain.CommonResponse[int]
// @Router /v1/dictionary/{id} [delete]
func (c *DictionaryDetailController) DeleteDictionary(ctx *gin.Context) {
	dictionaryID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		c.Logger.Error("Invalid dictionary ID parameter for deletion", zap.Error(err), zap.String("id", ctx.Param("id")))
		appError := domainErrors.NewAppError(errors.New("param id is necessary"), domainErrors.ValidationError)
		_ = ctx.Error(appError)
		return
	}
	c.Logger.Info("Deleting dictionary", zap.Int("id", dictionaryID))
	err = c.dictionaryService.Delete([]int{dictionaryID})
	if err != nil {
		c.Logger.Error("Error deleting dictionary", zap.Error(err), zap.Int("id", dictionaryID))
		_ = ctx.Error(err)
		return
	}
	c.Logger.Info("Dictionary deleted successfully", zap.Int("id", dictionaryID))
	ctx.JSON(http.StatusOK, domain.CommonResponse[int]{
		Data:    dictionaryID,
		Message: "resource deleted successfully",
		Status:  0,
	})
}

// SearchDictionaryPageList
// @Summary search dictionaries
// @Description search dictionaries by query
// @Tags search dictionaries
// @Accept json
// @Produce json
// @Success 200 {object} domain.PageList[[]ResponseDictionary]
// @Router /v1/dictionary/search [get]
func (c *DictionaryDetailController) SearchPaginated(ctx *gin.Context) {
	c.Logger.Info("Searching dictionaries with pagination")

	// Parse query parameters
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("pageSize", "10"))
	if pageSize < 1 {
		pageSize = 10
	}

	// Build filters
	filters := domain.DataFilters{
		Page:     page,
		PageSize: pageSize,
	}

	// Parse like filters
	likeFilters := make(map[string][]string)
	for field := range dictionaryRepo.ColumnsDictionaryMapping {
		if values := ctx.QueryArray(field + "_like"); len(values) > 0 {
			likeFilters[field] = values
		}
	}
	filters.LikeFilters = likeFilters

	// Parse exact matches
	matches := make(map[string][]string)
	for field := range dictionaryRepo.ColumnsDictionaryMapping {
		if values := ctx.QueryArray(field + "_match"); len(values) > 0 {
			matches[field] = values
		}
	}
	fmt.Println(matches)
	filters.Matches = matches

	// Parse date range filters
	var dateRanges []domain.DateRangeFilter
	for field := range dictionaryRepo.ColumnsDictionaryMapping {
		startStr := ctx.Query(field + "_start")
		endStr := ctx.Query(field + "_end")

		if startStr != "" || endStr != "" {
			dateRange := domain.DateRangeFilter{Field: field}

			if startStr != "" {
				if startTime, err := time.Parse(time.RFC3339, startStr); err == nil {
					dateRange.Start = &startTime
				}
			}

			if endStr != "" {
				if endTime, err := time.Parse(time.RFC3339, endStr); err == nil {
					dateRange.End = &endTime
				}
			}

			dateRanges = append(dateRanges, dateRange)
		}
	}
	filters.DateRangeFilters = dateRanges

	// Parse sorting
	sortBy := ctx.QueryArray("sortBy")
	if len(sortBy) > 0 {
		filters.SortBy = sortBy
	} else {
		filters.SortBy = []string{}
	}

	sortDirection := domain.SortDirection(ctx.DefaultQuery("sortDirection", "asc"))
	if sortDirection.IsValid() {
		filters.SortDirection = sortDirection
	}

	result, err := c.dictionaryService.SearchPaginated(filters)
	if err != nil {
		c.Logger.Error("Error searching dictionaries", zap.Error(err))
		_ = ctx.Error(err)
		return
	}
	type PageResult = domain.PageList[*[]*ResponseDictionary]
	response := controllers.NewCommonResponseBuilder[PageResult]().
		Data(PageResult{
			List:       arrayDomainToResponseMapper(result.Data),
			Total:      result.Total,
			Page:       result.Page,
			PageSize:   result.PageSize,
			TotalPages: result.TotalPages,
			Filters:    filters,
		}).
		Message("success").
		Status(0).
		Build()

	c.Logger.Info("Successfully searched dictionaries",
		zap.Int64("total", result.Total),
		zap.Int("page", result.Page))
	ctx.JSON(http.StatusOK, response)
}

// SearchByProperty
// @Summary  search by property
// @Description search by property
// @Tags search
// @Accept json
// @Produce json
// @Success 200 {array} []string
// @Router /v1/dictionary/search-property [get]
func (c *DictionaryDetailController) SearchByProperty(ctx *gin.Context) {
	property := ctx.Query("property")
	searchText := ctx.Query("searchText")

	if property == "" || searchText == "" {
		c.Logger.Error("Missing property or searchText parameter")
		appError := domainErrors.NewAppError(errors.New("missing property or searchText parameter"), domainErrors.ValidationError)
		_ = ctx.Error(appError)
		return
	}

	// Validate property
	allowed := map[string]bool{
		"dictionaryName": true,
		"email":          true,
		"firstName":      true,
		"lastName":       true,
		"status":         true,
	}
	if !allowed[property] {
		c.Logger.Error("Invalid property for search", zap.String("property", property))
		appError := domainErrors.NewAppError(errors.New("invalid property"), domainErrors.ValidationError)
		_ = ctx.Error(appError)
		return
	}

	coincidences, err := c.dictionaryService.SearchByProperty(property, searchText)
	if err != nil {
		c.Logger.Error("Error searching by property", zap.Error(err), zap.String("property", property))
		_ = ctx.Error(err)
		return
	}

	c.Logger.Info("Successfully searched by property",
		zap.String("property", property),
		zap.Int("results", len(*coincidences)))
	ctx.JSON(http.StatusOK, coincidences)
}

// BatchDeleteOperation
// @Summary delete operations
// @Description delete operations by id
// @Tags batch delete
// @Accept json
// @Produce json
// @Param book body DeleteBatchOperationRequest true  "JSON Data"
// @Success 200 {object} domain.CommonResponse[int]
// @Router /v1/operation/delete-batch [post]
func (c *DictionaryDetailController) DeleteDictionaryDetails(ctx *gin.Context) {
	c.Logger.Info("Creating new dictionary")
	var request DeleteBatchDictionaryDetailRequest
	var err error
	if err = controllers.BindJSON(ctx, &request); err != nil {
		c.Logger.Error("Error binding JSON for new dictionary", zap.Error(err))
		appError := domainErrors.NewAppError(err, domainErrors.ValidationError)
		_ = ctx.Error(appError)
		return
	}
	c.Logger.Info("Deleting operation", zap.String("ids", fmt.Sprintf("%v", request.IDS)))
	err = c.dictionaryService.Delete(request.IDS)
	if err != nil {
		c.Logger.Error("Error deleting operation", zap.Error(err), zap.String("ids", fmt.Sprintf("%v", request.IDS)))
		_ = ctx.Error(err)
		return
	}
	c.Logger.Info("Operation deleted successfully", zap.String("ids", fmt.Sprintf("%v", request.IDS)))
	ctx.JSON(http.StatusOK, domain.CommonResponse[[]int]{
		Data:    request.IDS,
		Message: "resource deleted successfully",
		Status:  0,
	})
}

// Mappers
func domainToResponseMapper(domainDictionary *domainDictionary.Dictionary) *ResponseDictionary {

	return &ResponseDictionary{
		ID:              domainDictionary.ID,
		Label:           domainDictionary.Label,
		Value:           domainDictionary.Value,
		Extend:          domainDictionary.Extend,
		Status:          domainDictionary.Status,
		Sort:            domainDictionary.Sort,
		SysDictionaryID: domainDictionary.SysDictionaryID,
		CreatedAt:       domain.CustomTime{Time: domainDictionary.CreatedAt},
		UpdatedAt:       domain.CustomTime{Time: domainDictionary.UpdatedAt},
	}
}

func arrayDomainToResponseMapper(dictionaries *[]domainDictionary.Dictionary) *[]*ResponseDictionary {
	res := make([]*ResponseDictionary, len(*dictionaries))
	for i, u := range *dictionaries {
		res[i] = domainToResponseMapper(&u)
	}
	return &res
}

func toUsecaseMapper(req *NewDictionaryRequest) *domainDictionary.Dictionary {
	return &domainDictionary.Dictionary{
		Label:           req.Label,
		Value:           req.Value,
		Extend:          req.Extend,
		Status:          req.Status,
		Sort:            req.Sort,
		SysDictionaryID: req.SysDictionaryID,
	}
}
