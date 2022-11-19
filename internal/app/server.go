package app

import (
	"context"
	"encoding/gob"
	"fmt"
	"net/http"

	"goquizbox/internal/middleware"
	"goquizbox/internal/repo/model"
	"goquizbox/internal/serverenv"
	"goquizbox/internal/web/auth"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
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
	tmpl, err := s.config.TemplateRenderer()
	if err != nil {
		panic(fmt.Errorf("failed to load templates: %w", err))
	}

	gob.Register(model.NewUser())

	mux := gin.New()

	// Recovery middleware recovers from any panics and writes a 500 if there was one.
	mux.Use(gin.Recovery())

	sessionStore := cookie.NewStore([]byte("supersecret"))       // To set supersecret from env vars
	sessionStore.Options(sessions.Options{MaxAge: 60 * 60 * 12}) // expire in 12 hrs

	mux.Use(sessions.Sessions("session", sessionStore))

	sessionAuthenticator := auth.NewSessionAuthenticator(s.env)

	defaultMiddlewares := middleware.DefaultMiddlewares(sessionAuthenticator)
	mux.Use(defaultMiddlewares...)

	mux.SetFuncMap(TemplateFuncMap)
	mux.SetHTMLTemplate(tmpl)

	// Static assets.
	mux.StaticFS("/assets/", http.FS(assetsFS))

	publicRoutes := mux.Group("/")
	publicRoutes.Use(middleware.MustNotBeLoggedIn)

	privateRoutes := mux.Group("/")
	privateRoutes.Use(middleware.MustBeLoggedIn)

	// Landing page
	publicRoutes.GET("/", s.HandleIndex())

	// Healthz page
	mux.GET("/healthz", s.HandleHealthz())
	mux.HEAD("/healthz", s.HandleHealthz())

	// Login pages
	publicRoutes.GET("/login", s.HandleLoginShow())
	publicRoutes.POST("/login", s.HandleLoginProcess())
	privateRoutes.GET("/logout", s.HandleLogout())
	privateRoutes.GET("/user-profile", s.HandleShowUserProfile())

	// Register pages
	publicRoutes.GET("/register", s.HandleRegisterShow())
	publicRoutes.POST("/register", s.HandleRegisterProcess())

	apiRoutes := mux.Group("/api/v1")
	{
		apiRoutes.POST("users", s.HandleAPIRegister())
		apiRoutes.POST("users/login", s.HandleAPILogin(sessionAuthenticator))

		securedApiRoutes := apiRoutes.Group("")
		securedApiRoutes.Use(auth.AllowOnlyActiveUser(
			sessionAuthenticator,
			s.env,
		))
		{
			securedApiRoutes.GET("/users", s.HandleApiListUsers())
			securedApiRoutes.GET("/users/:id", s.HandleApiGetUser())
			securedApiRoutes.PUT("/users/:id", s.HandleApiUpdateUser())
			securedApiRoutes.DELETE("/users/:id", s.HandleApiDeleteUser())
			securedApiRoutes.DELETE("/users/logout", s.HandleApiLogoutUser())
		}
	}

	mux.NoRoute(func(c *gin.Context) {
		c.String(http.StatusNotFound, "Not found")
	})
	return mux
}
