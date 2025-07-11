package di

import (
	"os"
	"testing"

	"github.com/gbrayhan/microservices-go/src/domain"
	domainMedicine "github.com/gbrayhan/microservices-go/src/domain/medicine"
	domainUser "github.com/gbrayhan/microservices-go/src/domain/user"
	logger "github.com/gbrayhan/microservices-go/src/infrastructure/logger"
	"github.com/gbrayhan/microservices-go/src/infrastructure/security"
	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock repositories and services
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) GetAll() (*[]domainUser.User, error) {
	args := m.Called()
	return args.Get(0).(*[]domainUser.User), args.Error(1)
}

func (m *MockUserRepository) GetByID(id int) (*domainUser.User, error) {
	args := m.Called(id)
	return args.Get(0).(*domainUser.User), args.Error(1)
}

func (m *MockUserRepository) Create(user *domainUser.User) (*domainUser.User, error) {
	args := m.Called(user)
	return args.Get(0).(*domainUser.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(email string) (*domainUser.User, error) {
	args := m.Called(email)
	return args.Get(0).(*domainUser.User), args.Error(1)
}
func (m *MockUserRepository) GetByUsername(username string) (*domainUser.User, error) {
	args := m.Called(username)
	return args.Get(0).(*domainUser.User), args.Error(1)
}
func (m *MockUserRepository) Delete(id int) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockUserRepository) Update(id int, userMap map[string]interface{}) (*domainUser.User, error) {
	args := m.Called(id, userMap)
	return args.Get(0).(*domainUser.User), args.Error(1)
}

func (m *MockUserRepository) SearchPaginated(filters domain.DataFilters) (*domainUser.SearchResultUser, error) {
	args := m.Called(filters)
	return args.Get(0).(*domainUser.SearchResultUser), args.Error(1)
}

func (m *MockUserRepository) SearchByProperty(property string, searchText string) (*[]string, error) {
	args := m.Called(property, searchText)
	return args.Get(0).(*[]string), args.Error(1)
}

func (m *MockUserRepository) GetOneByMap(userMap map[string]interface{}) (*domainUser.User, error) {
	args := m.Called(userMap)
	return args.Get(0).(*domainUser.User), args.Error(1)
}

type MockMedicineRepository struct {
	mock.Mock
}

func (m *MockMedicineRepository) GetAll() (*[]domainMedicine.Medicine, error) {
	args := m.Called()
	return args.Get(0).(*[]domainMedicine.Medicine), args.Error(1)
}

func (m *MockMedicineRepository) GetData(page int64, limit int64, sortBy string, sortDirection string, filters map[string][]string, searchText string, dateRangeFilters []domain.DateRangeFilter) (*domainMedicine.DataMedicine, error) {
	args := m.Called(page, limit, sortBy, sortDirection, filters, searchText, dateRangeFilters)
	return args.Get(0).(*domainMedicine.DataMedicine), args.Error(1)
}

func (m *MockMedicineRepository) GetByID(id int) (*domainMedicine.Medicine, error) {
	args := m.Called(id)
	return args.Get(0).(*domainMedicine.Medicine), args.Error(1)
}

func (m *MockMedicineRepository) Create(medicine *domainMedicine.Medicine) (*domainMedicine.Medicine, error) {
	args := m.Called(medicine)
	return args.Get(0).(*domainMedicine.Medicine), args.Error(1)
}

func (m *MockMedicineRepository) GetByMap(medicineMap map[string]any) (*domainMedicine.Medicine, error) {
	args := m.Called(medicineMap)
	return args.Get(0).(*domainMedicine.Medicine), args.Error(1)
}

func (m *MockMedicineRepository) Delete(id int) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockMedicineRepository) Update(id int, medicineMap map[string]any) (*domainMedicine.Medicine, error) {
	args := m.Called(id, medicineMap)
	return args.Get(0).(*domainMedicine.Medicine), args.Error(1)
}

func (m *MockMedicineRepository) SearchPaginated(filters domain.DataFilters) (*domainMedicine.SearchResultMedicine, error) {
	args := m.Called(filters)
	return args.Get(0).(*domainMedicine.SearchResultMedicine), args.Error(1)
}

func (m *MockMedicineRepository) SearchByProperty(property string, searchText string) (*[]string, error) {
	args := m.Called(property, searchText)
	return args.Get(0).(*[]string), args.Error(1)
}

type MockJWTService struct {
	mock.Mock
}

func (m *MockJWTService) GenerateJWTToken(userID int, tokenType string) (*security.AppToken, error) {
	args := m.Called(userID, tokenType)
	return args.Get(0).(*security.AppToken), args.Error(1)
}

func (m *MockJWTService) GetClaimsAndVerifyToken(tokenString string, tokenType string) (jwt.MapClaims, error) {
	args := m.Called(tokenString, tokenType)
	return args.Get(0).(jwt.MapClaims), args.Error(1)
}

type MockJwtBlackListService struct {
	mock.Mock
}

// AddToBlacklist(jwtToken string) error
// IsJwtInBlacklist(token string) (bool, error)
func (m *MockJwtBlackListService) AddToBlacklist(jwtToken string) error {
	args := m.Called(jwtToken)
	return args.Error(0)
}

func (m *MockJwtBlackListService) IsJwtInBlacklist(token string) (bool, error) {
	args := m.Called(token)
	return args.Get(0).(bool), args.Error(1)
}

func setupLogger(t *testing.T) *logger.Logger {
	loggerInstance, err := logger.NewLogger()
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	return loggerInstance
}

func TestNewTestApplicationContext(t *testing.T) {
	mockUserRepo := &MockUserRepository{}
	mockMedicineRepo := &MockMedicineRepository{}
	mockJWTService := &MockJWTService{}
	logger := setupLogger(t)
	jwtBlackListRepo := &MockJwtBlackListService{}

	appContext := NewTestApplicationContext(mockUserRepo, mockMedicineRepo, mockJWTService, logger, jwtBlackListRepo)

	assert.NotNil(t, appContext)
	assert.Equal(t, mockUserRepo, appContext.UserRepository)
	assert.Equal(t, mockMedicineRepo, appContext.MedicineRepository)
	assert.Equal(t, mockJWTService, appContext.JWTService)

	// Test that controllers are created
	assert.NotNil(t, appContext.AuthController)
	assert.NotNil(t, appContext.UserController)
	assert.NotNil(t, appContext.MedicineController)

	// Test that use cases are created
	assert.NotNil(t, appContext.AuthUseCase)
	assert.NotNil(t, appContext.UserUseCase)
	assert.NotNil(t, appContext.MedicineUseCase)
}

func TestSetupDependencies(t *testing.T) {
	// This test will fail in CI/CD without a real database connection
	// We'll test the error path by setting invalid environment variables
	originalPort := os.Getenv("DB_PORT")
	os.Setenv("DB_PORT", "99999") // Invalid port to cause connection failure
	defer os.Setenv("DB_PORT", originalPort)

	logger := setupLogger(t)
	appContext, err := SetupDependencies(logger)

	assert.Error(t, err)
	assert.Nil(t, appContext)
}

func TestApplicationContextStructure(t *testing.T) {
	mockUserRepo := &MockUserRepository{}
	mockMedicineRepo := &MockMedicineRepository{}
	mockJWTService := &MockJWTService{}
	logger := setupLogger(t)
	jwtBlackListRepo := &MockJwtBlackListService{}
	appContext := NewTestApplicationContext(mockUserRepo, mockMedicineRepo, mockJWTService, logger, jwtBlackListRepo)

	// Test that all fields are properly set
	assert.NotNil(t, appContext.AuthController)
	assert.NotNil(t, appContext.UserController)
	assert.NotNil(t, appContext.MedicineController)
	assert.NotNil(t, appContext.JWTService)
	assert.NotNil(t, appContext.UserRepository)
	assert.NotNil(t, appContext.MedicineRepository)
	assert.NotNil(t, appContext.AuthUseCase)
	assert.NotNil(t, appContext.UserUseCase)
	assert.NotNil(t, appContext.MedicineUseCase)

	// Test that DB is nil in test context (as expected)
	assert.Nil(t, appContext.DB)
}
