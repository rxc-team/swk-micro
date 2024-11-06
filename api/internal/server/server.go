package server

import (
	"github.com/gin-gonic/gin"

	"rxcsoft.cn/pit3/api/internal/middleware/log"
)

type (
	// Engine server engine struct
	Engine struct {
		GinEngine *gin.Engine
	}
)

var (
	ginEngine *gin.Engine
)

// NewWebHost new gin webhost
func NewWebHost() *Engine {
	s := Engine{
		GinEngine: gin.New(),
	}
	return &s
}

// UseExternalAPIRoutes set external API routes
func (s *Engine) UseExternalAPIRoutes(startAPPAPIRoutes func(*gin.Engine) error) *Engine {
	startAPPAPIRoutes(s.GinEngine)

	return s
}

// UseGinRecovery use gin http recovery
func (s *Engine) UseGinRecovery() *Engine {
	(*s).GinEngine.Use(gin.Recovery())
	return s
}

// UseGinLogger use gin default logger
func (s *Engine) UseGinLogger() *Engine {
	// (*s.GinEngine).Use(gin.Logger())
	(*s.GinEngine).Use(log.Logger())
	return s
}

// UseMiddlewares add custom middlewares to gin
func (s *Engine) UseMiddlewares(middlewares ...gin.HandlerFunc) *Engine {
	for _, middleware := range middlewares {
		(*s).GinEngine.Use(middleware)
	}

	return s
}

// GetGinEngine get current gin engine
func (s *Engine) GetGinEngine() *gin.Engine {
	return ginEngine
}

// Run run the server stack
func (s *Engine) Run(port string) *Engine {
	(*s).GinEngine.Run(port)

	return s
}

// DatabaseStartups initialize database
func (s *Engine) DatabaseStartups(dbs ...func()) *Engine {
	for _, startdb := range dbs {
		startdb()
	}

	return s
}

// ConfigureServices configure services from func
func (s *Engine) ConfigureServices(services func()) *Engine {
	services()

	return s
}
