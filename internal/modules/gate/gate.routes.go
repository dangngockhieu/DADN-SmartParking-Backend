package gate

import "github.com/gin-gonic/gin"

func RegisterRoutes(api *gin.RouterGroup, handler *Handler, authMiddleware, adminOnly gin.HandlerFunc) {
	// /gates
	gates := api.Group("/gates")
	gates.Use(authMiddleware, adminOnly)
	{
		gates.POST("", handler.Create)
		gates.GET("/:gateId", handler.GetByID)
		gates.PUT("/:gateId", handler.Update)
		gates.DELETE("/:gateId", handler.Delete)
	}
}
