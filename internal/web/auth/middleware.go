package auth

import (
	"fmt"
	"net/http"

	"goquizbox/internal/repo/database"
	"goquizbox/internal/serverenv"
	"goquizbox/internal/web/ctxhelper"
	"goquizbox/pkg/logging"

	"github.com/gin-gonic/gin"
)

func AllowOnlyActiveUser(
	sessionAuthenticator SessionAuthenticator,
	env *serverenv.ServerEnv,
) func(c *gin.Context) {

	return func(c *gin.Context) {

		ctx := c.Request.Context()
		logger := logging.FromContext(ctx).Named("allowOnlyActiveUser")

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
		ctx := c.Request.Context()
		logger := logging.FromContext(ctx).Named("allowWithSession")

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

	sessionDB := database.NewSessionDB(env.Database())
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

// func AllowOnlyValidApiKey(
// 	dB db.DB,
// 	apiKeyService services.ApiKeyService,
// ) func(c *gin.Context) {

// 	return func(c *gin.Context) {

// 		strategy := "basic_auth"
// 		var username, password string

// 		auth := strings.SplitN(c.Request.Header.Get("Authorization"), " ", 2)
// 		if len(auth) != 2 || auth[0] != "Basic" {
// 			strategy = "url_params"
// 		}

// 		switch strategy {
// 		case "basic_auth":
// 			payload, err := base64.StdEncoding.DecodeString(auth[1])
// 			if err != nil {
// 				respondWithError(c, 401, "Unauthorized. Basic Authentication required.")
// 				return
// 			}

// 			pair := strings.SplitN(string(payload), ":", 2)
// 			if len(pair) != 2 {
// 				respondWithError(c, 401, "Unauthorized. Basic Authentication required, given auth not correct.")
// 				return
// 			}

// 			username = pair[0]
// 			password = pair[1]
// 		case "url_params":
// 			username = c.Query("username")
// 			password = c.Query("password")

// 			if username == "" || password == "" {
// 				respondWithError(c, 401, "Unauthorized. Authentication required.")
// 				return
// 			}

// 		default:
// 			respondWithError(c, 401, "Unauthorized. Authentication required.")
// 			return
// 		}

// 		ctx := c.Request.Context()
// 		cachedApiKey, err := apiKeyService.ValidateApiKey(ctx, dB, username, password)
// 		if err != nil {
// 			logger.Errorf("Failed to validate api key: %v", err)
// 			respondWithError(c, 500, "Unable to perform request")
// 			return
// 		}

// 		ctx = ctxhelper.WithTokenInfo(ctx, &entities.TokenInfo{UserID: cachedApiKey.UserID})
// 		c.Request = c.Request.WithContext(ctx)
// 		c.Next()
// 	}
// }

func respondWithError(c *gin.Context, code int, message string) {
	c.JSON(code, map[string]interface{}{"success": false, "message": message})
	c.Abort()
}
