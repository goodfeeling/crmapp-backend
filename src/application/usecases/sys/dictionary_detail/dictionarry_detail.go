package dictionary_detail

import (
	"fmt"

	logger "github.com/gbrayhan/microservices-go/src/infrastructure/logger"
	dictionaryRepo "github.com/gbrayhan/microservices-go/src/infrastructure/repository/psql/sys/dictionary_detail"

	"github.com/gbrayhan/microservices-go/src/domain"
	dictionaryDomain "github.com/gbrayhan/microservices-go/src/domain/sys/dictionary_detail"
	"go.uber.org/zap"
)

type ISysDictionaryService interface {
	GetAll() (*[]dictionaryDomain.Dictionary, error)
	GetByID(id int) (*dictionaryDomain.Dictionary, error)
	Create(newDictionary *dictionaryDomain.Dictionary) (*dictionaryDomain.Dictionary, error)
	Delete(ids []int) error
	Update(id int, userMap map[string]interface{}) (*dictionaryDomain.Dictionary, error)
	SearchPaginated(filters domain.DataFilters) (*domain.PaginatedResult[dictionaryDomain.Dictionary], error)
	SearchByProperty(property string, searchText string) (*[]string, error)
	GetOneByMap(userMap map[string]interface{}) (*dictionaryDomain.Dictionary, error)
}

type SysDictionaryUseCase struct {
	sysDictionaryRepository dictionaryRepo.DictionaryRepositoryInterface
	Logger                  *logger.Logger
}

func NewSysDictionaryUseCase(sysDictionaryRepository dictionaryRepo.DictionaryRepositoryInterface, loggerInstance *logger.Logger) ISysDictionaryService {
	return &SysDictionaryUseCase{
		sysDictionaryRepository: sysDictionaryRepository,
		Logger:                  loggerInstance,
	}
}

func (s *SysDictionaryUseCase) GetAll() (*[]dictionaryDomain.Dictionary, error) {
	s.Logger.Info("Getting all roles")
	return s.sysDictionaryRepository.GetAll()
}

func (s *SysDictionaryUseCase) GetByID(id int) (*dictionaryDomain.Dictionary, error) {
	s.Logger.Info("Getting dictionary by ID", zap.Int("id", id))
	return s.sysDictionaryRepository.GetByID(id)
}

func (s *SysDictionaryUseCase) Create(newDictionary *dictionaryDomain.Dictionary) (*dictionaryDomain.Dictionary, error) {
	s.Logger.Info("Creating new dictionary", zap.String("Label", newDictionary.Label))
	return s.sysDictionaryRepository.Create(newDictionary)
}

func (s *SysDictionaryUseCase) Delete(ids []int) error {
	s.Logger.Info("Deleting dictionary", zap.String("ids", fmt.Sprintf("%v", ids)))
	return s.sysDictionaryRepository.Delete(ids)
}

func (s *SysDictionaryUseCase) Update(id int, userMap map[string]interface{}) (*dictionaryDomain.Dictionary, error) {
	s.Logger.Info("Updating dictionary", zap.Int("id", id))
	return s.sysDictionaryRepository.Update(id, userMap)
}

func (s *SysDictionaryUseCase) SearchPaginated(filters domain.DataFilters) (*domain.PaginatedResult[dictionaryDomain.Dictionary], error) {
	s.Logger.Info("Searching dictionary with pagination",
		zap.Int("page", filters.Page),
		zap.Int("pageSize", filters.PageSize))
	return s.sysDictionaryRepository.SearchPaginated(filters)
}

func (s *SysDictionaryUseCase) SearchByProperty(property string, searchText string) (*[]string, error) {
	s.Logger.Info("Searching dictionary by property",
		zap.String("property", property),
		zap.String("searchText", searchText))
	return s.sysDictionaryRepository.SearchByProperty(property, searchText)
}

func (s *SysDictionaryUseCase) GetOneByMap(userMap map[string]interface{}) (*dictionaryDomain.Dictionary, error) {
	return s.sysDictionaryRepository.GetOneByMap(userMap)
}
