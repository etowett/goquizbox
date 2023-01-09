package app

import (
	"context"
	"fmt"
	"net/http"

	"goquizbox/internal/middleware"
	"goquizbox/internal/serverenv"
	"goquizbox/internal/web/auth"

	"github.com/gin-gonic/gin"
)

type Server struct {
	config *Config
	env    *serverenv.ServerEnv
}

func NewServer(config *Config, env *serverenv.ServerEnv) (*Server, error) {
	if env.Database() == nil {
		return nil, fmt.Errorf("missing Database in server env")
	}

	return &Server{
		config: config,
		env:    env,
	}, nil
}

func (s *Server) Routes(ctx context.Context) http.Handler {
	mux := gin.New()

	// Recovery middleware recovers from any panics and writes a 500 if there was one.
	mux.Use(gin.Recovery())

	sessionAuthenticator := auth.NewSessionAuthenticator(s.env)

	defaultMiddlewares := middleware.DefaultMiddlewares(sessionAuthenticator)
	mux.Use(defaultMiddlewares...)

	// Healthz page
	mux.GET("/healthz", s.HandleHealthz())

	apiRoutes := mux.Group("/api/v1")
	{
		apiRoutes.POST("users", s.HandleRegister())
		apiRoutes.POST("auth/login", s.HandleLogin(sessionAuthenticator))
		apiRoutes.GET("/users", s.HandleListUsers())
		apiRoutes.GET("/users/:id", s.HandleGetUser())

		apiRoutes.GET("/questions", s.HandleListQuestions())
		apiRoutes.GET("/questions/:id", s.HandleApiGetQuestion())
		apiRoutes.GET("/questions/:id/answers", s.HandleApiGetQuestionAnswers())

		securedApiRoutes := apiRoutes.Group("")
		securedApiRoutes.Use(auth.AllowOnlyActiveUser(
			sessionAuthenticator,
			s.env,
		))
		{
			securedApiRoutes.PUT("/users/:id", s.HandleApiUpdateUser())
			securedApiRoutes.DELETE("/users/:id", s.HandleApiDeleteUser())
			securedApiRoutes.DELETE("/auth/logout", s.HandleApiLogoutUser())

			securedApiRoutes.POST("/questions", s.HandleApiAddQuestion())
			securedApiRoutes.POST("/questions/:id/answers", s.HandleApiAddQuestionAnswer())
		}
	}

	mux.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "Not Found",
		})
	})
	return mux
}
