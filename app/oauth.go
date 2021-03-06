package app

import (
	"context"
	"github.com/deissh/go-utils"
	"github.com/dgrijalva/jwt-go"
	"github.com/rl-os/api/entity"
	"github.com/rl-os/api/entity/request"
	"github.com/rl-os/api/errors"
	"github.com/rl-os/api/repository"
	"net/http"
	"time"
)

var (
	// OAuthGroundType содержит доступные способы авторизации через oauth2
	// password - авторизация с помощью логина и
	// пароля которые передаются в теле запроса
	// refresh_token - авторизация с использованием
	// рефреш токена который был получен во время предыдущей авторизации
	OAuthGroundType = []string{"password", "refresh_token"}

	// Список ошибок которые может вернуть данный модуль
	// необходим для того что бы не явно не повторять
	// уникальный индификатор ошибки

	InvalidAuthParamsErr        = errors.New("invalid_auth_params", http.StatusBadRequest, "Not found")
	InvalidAuthTokenErr         = errors.New("oauth_validate_token", http.StatusUnauthorized, "Invalid access token")
	ExpiredAuthTokenErr         = errors.New("oauth_expire_token", http.StatusUnauthorized, "Access token expired")
	NotFoundRefreshAuthErr      = errors.New("oauth_invalid_refresh", http.StatusBadRequest, "Not found")
	InvalidOAuthClientParamsErr = errors.New("oauth_invalid_client_params", http.StatusBadRequest, "Invalid oauth client")
)

type OAuthUseCase struct {
	*App
	OAuthRepository repository.OAuth
	UserRepository  repository.User
}

func NewOAuthUseCase(app *App, rep repository.OAuth, user repository.User) *OAuthUseCase {
	return &OAuthUseCase{app, rep, user}
}

// ValidateToken проверяет переданный функции токен доступа,
// если все нормально и токен еще не истек то вернется entity.OAuthToken
// при ошибке может вернуть InvalidAuthTokenErr или ExpiredAuthTokenErr
func (a *OAuthUseCase) ValidateToken(ctx context.Context, accessToken string) (*entity.OAuthToken, error) {
	_, err := jwt.Parse(accessToken, func(token *jwt.Token) (interface{}, error) {
		return []byte(a.Options.JWT.Secret), nil
	})

	if err != nil {
		return nil, InvalidAuthTokenErr.WithCause(err)
	}

	token, err := a.OAuthRepository.GetToken(ctx, accessToken)
	if err != nil {
		return nil, InvalidAuthTokenErr.WithCause(err, "Not found")
	}

	if token.Revoked {
		return nil, ExpiredAuthTokenErr
	}

	return token, nil
}

// CreateOAuthClient создает новый entity.OAuthClient
// в случае InvalidOAuthClientParamsErr
func (a *OAuthUseCase) CreateOAuthClient(ctx context.Context, userId uint, request request.CreateOAuthClient) (*entity.OAuthClient, error) {
	if err := a.Validator.Struct(&request); err != nil {
		return nil, InvalidAuthParamsErr.WithCause(err)
	}

	token, err := a.OAuthRepository.CreateClient(ctx, userId, request.Name, request.Redirect)
	if err != nil {
		return nil, InvalidOAuthClientParamsErr.WithCause(err)
	}

	return token, nil
}

// RefreshToken and revoke old access token
func (a *OAuthUseCase) RefreshToken(ctx context.Context, refreshToken string) (*entity.OAuthToken, error) {
	token, err := a.OAuthRepository.RefreshToken(ctx, refreshToken)
	if err != nil {
		return nil, NotFoundRefreshAuthErr.WithCause(err)
	}

	return token, nil
}

// CreateOAuthToken and return it
func (a *OAuthUseCase) CreateOAuthToken(ctx context.Context, request request.CreateOauthToken) (*entity.OAuthToken, error) {
	if err := a.Validator.Struct(request); err != nil {
		return nil, InvalidAuthParamsErr.WithCause(err)
	}

	if !utils.ContainsString(OAuthGroundType, request.GrantType) {
		request.GrantType = "password"
	}

	if request.GrantType == "password" {
		user, err := a.UserRepository.GetByBasic(
			ctx,
			request.Username,
			request.Password,
		)
		if err != nil {
			return nil, InvalidAuthParamsErr.WithCause(err)
		}

		rawToken, _ := tokenCreate(
			a.Options.JWT.Secret,
			time.Hour*24,
			user.ID,
			request.ClientID,
			request.Scope,
		)

		token, err := a.OAuthRepository.CreateToken(ctx, rawToken)
		if err != nil {
			return nil, InvalidAuthParamsErr.WithCause(err)
		}

		return token, nil
	} else if request.GrantType == "refresh_token" {
		return a.RefreshToken(ctx, request.RefreshToken)
	}

	return nil, InvalidAuthParamsErr.WithCause(nil, "invalid grant type")
}
