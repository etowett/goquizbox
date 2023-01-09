package auth

import (
	"context"
	"errors"
	"fmt"
	"goquizbox/internal/entities"
	"goquizbox/internal/repos"
	"goquizbox/internal/serverenv"
	"net/http"
	"time"
)

const (
	tokenHeader = "X-Auth-Token"
)

var ErrTokenNotProvided = errors.New("token not provided")

type SessionAuthenticator interface {
	RefreshTokenFromRequest(ctx context.Context, tokenInfo *entities.TokenInfo, w http.ResponseWriter) (string, error)
	SetUserSessionInResponse(w http.ResponseWriter, user *entities.User, session *entities.Session) (string, error)
	TokenInfoFromRequest(req *http.Request) (*entities.TokenInfo, error)
	UserByID(ctx context.Context, userID int64) (*entities.User, error)
}

type AppSessionAuthenticator struct {
	jwtHandler JWTHandler
	env        *serverenv.ServerEnv
}

func NewSessionAuthenticator(env *serverenv.ServerEnv) SessionAuthenticator {
	return NewSessionAuthenticatorWithJWTHandler(
		NewJWTHandler(),
		env,
	)
}

func NewSessionAuthenticatorWithJWTHandler(
	jwtHandler JWTHandler,
	env *serverenv.ServerEnv,
) SessionAuthenticator {
	return &AppSessionAuthenticator{
		jwtHandler: jwtHandler,
		env:        env,
	}
}

func (a *AppSessionAuthenticator) RefreshTokenFromRequest(
	ctx context.Context,
	tokenInfo *entities.TokenInfo,
	w http.ResponseWriter,
) (string, error) {

	sessionDB := repos.NewSessionDB(a.env.Database())
	session, err := sessionDB.GetSessionByID(ctx, tokenInfo.SessionID)
	if err != nil {
		return "", fmt.Errorf("failed to find session by id=[%v]: %v", tokenInfo.SessionID, err)
	}

	if session.DeactivatedAt.Valid {
		return "", fmt.Errorf("failed to refresh session by id=[%v], deactivated ", tokenInfo.SessionID)
	}

	db := repos.NewUserDB(a.env.Database())
	user, err := db.GetByID(ctx, session.UserID)
	if err != nil {
		return "", fmt.Errorf("failed to find user by id=[%v]: %v", session.UserID, err)
	}

	if !user.Status.IsActive() {
		return "", fmt.Errorf("failed to refresh session by id=[%v] for user with id=[%v], user status inactive", tokenInfo.SessionID, session.UserID)
	}

	session.LastRefreshedAt = time.Now()

	err = sessionDB.Save(ctx, session)
	if err != nil {
		return "", fmt.Errorf("failed to update last refreshed at timestamp for session=[%v]: %v", session.ID, err)
	}

	return a.SetUserSessionInResponse(w, user, session)
}

func (a *AppSessionAuthenticator) SetUserSessionInResponse(
	w http.ResponseWriter,
	user *entities.User,
	session *entities.Session,
) (string, error) {

	tokenValue, err := a.jwtHandler.CreateUserToken(user, session)
	if err != nil {
		return "", fmt.Errorf("failed to generate signed user token for user=[%v]", user.ID)
	}

	w.Header().Set(tokenHeader, tokenValue)

	return tokenValue, nil
}

func (a *AppSessionAuthenticator) TokenInfoFromRequest(
	req *http.Request,
) (*entities.TokenInfo, error) {
	tokenValue := req.Header.Get(tokenHeader)
	if tokenValue != "" {
		return a.jwtHandler.TokenInfo(tokenValue)
	}

	return &entities.TokenInfo{}, ErrTokenNotProvided
}

func (a *AppSessionAuthenticator) UserByID(
	ctx context.Context,
	userID int64,
) (*entities.User, error) {

	db := repos.NewUserDB(a.env.Database())
	user, err := db.GetByID(ctx, userID)
	if err != nil {
		return user, fmt.Errorf("failed to find user by id=[%v]: %v", userID, err)
	}

	return user, nil
}
