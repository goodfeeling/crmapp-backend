package auth

import (
	"errors"
	"net/http"
	"strings"

	useCaseAuth "github.com/gbrayhan/microservices-go/src/application/usecases/auth"
	domain "github.com/gbrayhan/microservices-go/src/domain"
	domainErrors "github.com/gbrayhan/microservices-go/src/domain/errors"
	logger "github.com/gbrayhan/microservices-go/src/infrastructure/logger"
	"github.com/gbrayhan/microservices-go/src/infrastructure/rest/controllers"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type IAuthController interface {
	Login(ctx *gin.Context)
	Logout(ctx *gin.Context)
	Register(ctx *gin.Context)
	GetAccessTokenByRefreshToken(ctx *gin.Context)
}

type AuthController struct {
	authUseCase useCaseAuth.IAuthUseCase
	Logger      *logger.Logger
}

// RegisterUser godoc
//
//	@Summary		register new user
//	@Description	register new user
//	@Tags			register user
//	@Accept			json
//	@Produce		json
//	@Param			book	body	RegisterRequest	true	"JSON Data"
//	@Success		200		{array}	domain.CommonResponse[useCaseAuth.SecurityRegisterUser]
//	@Router			/v1/auth/signup [get]
func (c *AuthController) Register(ctx *gin.Context) {
	var request RegisterRequest
	if err := controllers.BindJSON(ctx, &request); err != nil {
		appError := domainErrors.NewAppError(err, domainErrors.ValidationError)
		_ = ctx.Error(appError)
		return
	}
	userRegister := useCaseAuth.RegisterUser{
		UserName: request.UserName,
		Email:    request.Email,
		Password: request.Password,
	}

	registerUser, err := c.authUseCase.Register(userRegister)
	if err != nil {
		appError := domainErrors.NewAppError(err, domainErrors.ValidationError)
		_ = ctx.Error(appError)
		return
	}
	ctx.JSON(http.StatusOK, registerUser)
}

// UserLogout godoc
// @Summary user logout
// @Description user logout
// @Tags logout
// @Accept json
// @Produce json
// @Success 200 {object} domain.CommonResponse[string]
// @Router /v1/auth/logout [get]
func (c *AuthController) Logout(ctx *gin.Context) {
	rawtoken := ctx.Request.Header.Get("Authorization")
	tokens := strings.Split(rawtoken, " ")
	if len(tokens) < 2 {
		appError := domainErrors.NewAppError(errors.New("token error"), domainErrors.ValidationError)
		_ = ctx.Error(appError)
		return
	}
	res, err := c.authUseCase.Logout(tokens[1])
	if err != nil {
		appError := domainErrors.NewAppError(err, domainErrors.ValidationError)
		_ = ctx.Error(appError)
		return
	}
	ctx.JSON(http.StatusOK, res)
}

func NewAuthController(authUsecase useCaseAuth.IAuthUseCase, loggerInstance *logger.Logger) IAuthController {
	return &AuthController{
		authUseCase: authUsecase,
		Logger:      loggerInstance,
	}
}

// Login godoc
// @Summary login godoc
// @Description login
// @Tags login
// @Accept json
// @Produce json
// @Param book body LoginRequest  true  "JSON Data"
// @Success 200 {object} domain.CommonResponse[useCaseAuth.SecurityAuthenticatedUser]
// @Router /v1/auth/signin [get]
func (c *AuthController) Login(ctx *gin.Context) {
	c.Logger.Info("User login request")
	var request LoginRequest
	if err := controllers.BindJSON(ctx, &request); err != nil {
		c.Logger.Error("Error binding JSON for login", zap.Error(err))
		appError := domainErrors.NewAppError(err, domainErrors.ValidationError)
		_ = ctx.Error(appError)
		return
	}

	domainUser, authTokens, err := c.authUseCase.Login(request.Username, request.Password)
	if err != nil {
		c.Logger.Error("Login failed", zap.Error(err), zap.String("email", request.Username))
		_ = ctx.Error(err)
		return
	}

	response := &domain.CommonResponse[useCaseAuth.SecurityAuthenticatedUser]{
		Data: useCaseAuth.SecurityAuthenticatedUser{
			UserInfo: useCaseAuth.DataUserAuthenticated{
				UserName:  domainUser.UserName,
				Email:     domainUser.Email,
				ID:        domainUser.ID,
				Status:    domainUser.Status,
				NickName:  domainUser.NickName,
				Phone:     domainUser.Phone,
				HeaderImg: domainUser.HeaderImg,
			},
			Security: useCaseAuth.DataSecurityAuthenticated{
				JWTAccessToken:            authTokens.AccessToken,
				JWTRefreshToken:           authTokens.RefreshToken,
				ExpirationAccessDateTime:  authTokens.ExpirationAccessDateTime,
				ExpirationRefreshDateTime: authTokens.ExpirationRefreshDateTime,
			},
		},
	}

	c.Logger.Info("Login successful", zap.String("email", request.Username), zap.Int("userID", domainUser.ID))
	ctx.JSON(http.StatusOK, response)
}

// RefreshUserToken
// @Summary refresh token
// @Description refresh token
// @Tags refresh_token
// @Accept json
// @Produce json
// @Param book body AccessTokenRequest  true  "JSON Data"
// @Success 200 {object} domain.CommonResponse[useCaseAuth.SecurityAuthenticatedUser]
// @Router /v1/auth/access-token [get]
func (c *AuthController) GetAccessTokenByRefreshToken(ctx *gin.Context) {
	c.Logger.Info("Token refresh request")
	var request AccessTokenRequest
	if err := controllers.BindJSON(ctx, &request); err != nil {
		c.Logger.Error("Error binding JSON for token refresh", zap.Error(err))
		appError := domainErrors.NewAppError(err, domainErrors.ValidationError)
		_ = ctx.Error(appError)
		return
	}

	domainUser, authTokens, err := c.authUseCase.AccessTokenByRefreshToken(request.RefreshToken)
	if err != nil {
		c.Logger.Error("Token refresh failed", zap.Error(err))
		appError := domainErrors.NewAppError(err, domainErrors.TokenExpired)
		_ = ctx.Error(appError)
		return
	}

	response := &domain.CommonResponse[useCaseAuth.SecurityAuthenticatedUser]{
		Data: useCaseAuth.SecurityAuthenticatedUser{
			UserInfo: useCaseAuth.DataUserAuthenticated{
				UserName:  domainUser.UserName,
				Email:     domainUser.Email,
				ID:        domainUser.ID,
				Status:    domainUser.Status,
				NickName:  domainUser.NickName,
				Phone:     domainUser.Phone,
				HeaderImg: domainUser.HeaderImg,
			},
			Security: useCaseAuth.DataSecurityAuthenticated{
				JWTAccessToken:            authTokens.AccessToken,
				JWTRefreshToken:           authTokens.RefreshToken,
				ExpirationAccessDateTime:  authTokens.ExpirationAccessDateTime,
				ExpirationRefreshDateTime: authTokens.ExpirationRefreshDateTime,
			},
		},
	}

	c.Logger.Info("Token refresh successful", zap.Int("userID", domainUser.ID))
	ctx.JSON(http.StatusOK, response)
}
