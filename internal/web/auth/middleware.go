package auth

import (
	"fmt"
	"net/http"

	"goquizbox/internal/logger"
	"goquizbox/internal/repos"
	"goquizbox/internal/serverenv"
	"goquizbox/internal/web/ctxhelper"

	"github.com/gin-gonic/gin"
)

func AllowOnlyActiveUser(
	sessionAuthenticator SessionAuthenticator,
	env *serverenv.ServerEnv,
) func(c *gin.Context) {

	return func(c *gin.Context) {

		ctx := c.Request.Context()

		err := validateSession(c, sessionAuthenticator, env)
		if err != nil {
			logger.Errorf("could not validate session: %v", err)
			c.JSON(http.StatusBadRequest, map[string]interface{}{
				"success": false,
				"message": "failed to validate session, you have to log in",
			})
			c.Abort()
			return
		}

		tokenInfo := ctxhelper.TokenInfo(ctx)

		if tokenInfo.RequiresRefresh() {
			_, err = sessionAuthenticator.RefreshTokenFromRequest(ctx, tokenInfo, c.Writer)
			if err != nil {
				logger.Errorf("could not refresh token from request: %v", err)

				c.JSON(http.StatusBadRequest, map[string]interface{}{
					"success": false,
					"message": fmt.Sprintf("Failed to refresh token from request: %v", err),
				})
				c.Abort()
				return
			}
		}
	}
}

func AllowWithSession(
	sessionAuthenticator SessionAuthenticator,
	env *serverenv.ServerEnv,
) func(c *gin.Context) {

	return func(c *gin.Context) {
		err := validateSession(c, sessionAuthenticator, env)
		if err != nil {
			logger.Errorf("could not validate session: %v", err)
			c.JSON(http.StatusBadRequest, map[string]interface{}{
				"success": false,
				"message": "failed to validate session, you have to log in",
			})
			c.Abort()
			return
		}
	}
}

func validateSession(
	c *gin.Context,
	sessionAuthenticator SessionAuthenticator,
	env *serverenv.ServerEnv,
) error {

	ctx := c.Request.Context()
	tokenInfo := ctxhelper.TokenInfo(ctx)

	// TODO: Add caching here to prevent db lookup each time

	sessionDB := repos.NewSessionDB(env.Database())
	session, err := sessionDB.GetFullSessionByID(ctx, tokenInfo.SessionID)
	if err != nil {
		return fmt.Errorf("failed to get full session for id %v - %v", tokenInfo.SessionID, err)
	}

	if session == nil {
		return fmt.Errorf("session not found")
	}

	if session.UserID != tokenInfo.UserID {
		return fmt.Errorf("check for sessionID=[%v] with userID=[%v], and user=[%v] not match", tokenInfo.SessionID, session.UserID, tokenInfo.UserID)
	}

	if !session.UserStatus.IsActive() {
		return fmt.Errorf("user=[%v] is inactive", tokenInfo.UserID)
	}

	if session.DeactivatedAt.Valid {
		return fmt.Errorf("sessionID=[%v] is not active", tokenInfo.SessionID)
	}

	return nil
}
