package parking_lot

import "github.com/gin-gonic/gin"

func RegisterRoutes(api *gin.RouterGroup, handler *Handler, authMiddleware, managerOrAdmin gin.HandlerFunc) {
	group := api.Group("/parking-lots")
	group.Use(authMiddleware)
	{
		group.POST("", managerOrAdmin, handler.Create)
		group.GET("", handler.FindAll)
		group.GET("/:id", handler.FindByID)
		group.GET("/:id/gates", managerOrAdmin, handler.GetGatesByLotID)
		group.PATCH("/:id", managerOrAdmin, handler.Update)
	}
}
