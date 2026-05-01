package parking_lot

import "github.com/gin-gonic/gin"

func RegisterRoutes(api *gin.RouterGroup, handler *Handler, authMiddleware, adminOnly gin.HandlerFunc) {
	group := api.Group("/parking-lots")
	group.Use(authMiddleware)
	{
		group.POST("", adminOnly, handler.Create)
		group.GET("", handler.FindAll)
		group.GET("/:id", handler.FindByID)
		group.GET("/:id/gates", adminOnly, handler.GetGatesByLotID)
		group.PATCH("/:id", adminOnly, handler.Update)
	}
}
