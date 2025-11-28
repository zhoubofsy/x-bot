package api

import (
	"github.com/gin-gonic/gin"
	"github.com/zhoubofsy/x-bot/internal/interfaces/api/handler"
	"github.com/zhoubofsy/x-bot/internal/interfaces/api/middleware"
)

type Router struct {
	engine          *gin.Engine
	workflowHandler *handler.WorkflowHandler
	adCopyHandler   *handler.AdCopyHandler
}

func NewRouter(
	workflowHandler *handler.WorkflowHandler,
	adCopyHandler *handler.AdCopyHandler,
	mode string,
	apiKey string,
) *Router {
	if mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.New()
	engine.Use(middleware.Logger())
	engine.Use(middleware.Recovery())

	r := &Router{
		engine:          engine,
		workflowHandler: workflowHandler,
		adCopyHandler:   adCopyHandler,
	}

	r.setupRoutes(apiKey)

	return r
}

func (r *Router) setupRoutes(apiKey string) {
	// Health check
	r.engine.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// API v1
	v1 := r.engine.Group("/api/v1")
	v1.Use(middleware.APIKeyAuth(apiKey))
	{
		// Workflow
		workflow := v1.Group("/workflow")
		{
			workflow.POST("/execute", r.workflowHandler.Execute)
			workflow.POST("/sync-following", r.workflowHandler.SyncFollowing)
		}

		// Stats & Logs
		v1.GET("/stats", r.workflowHandler.GetStats)
		v1.GET("/reply-logs", r.workflowHandler.GetRecentLogs)

		// Ad Copies
		adCopies := v1.Group("/ad-copies")
		{
			adCopies.GET("", r.adCopyHandler.List)
			adCopies.GET("/:id", r.adCopyHandler.Get)
			adCopies.POST("", r.adCopyHandler.Create)
			adCopies.PUT("/:id", r.adCopyHandler.Update)
			adCopies.DELETE("/:id", r.adCopyHandler.Delete)
		}
	}
}

func (r *Router) Engine() *gin.Engine {
	return r.engine
}

