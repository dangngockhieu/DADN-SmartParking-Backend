package gate

import "github.com/gin-gonic/gin"

func RegisterRoutes(api *gin.RouterGroup, handler *Handler, authMiddleware, managerOrAdmin gin.HandlerFunc) {
	// /gates
	gates := api.Group("/gates")
	gates.Use(authMiddleware, managerOrAdmin)
	{
		gates.POST("", handler.Create)
		gates.GET("/:gateId", handler.GetByID)
		gates.PUT("/:gateId", handler.Update)
		gates.DELETE("/:gateId", handler.Delete)
	}
}
