package parking_slot

import "github.com/gin-gonic/gin"

func RegisterRoutes(api *gin.RouterGroup, handler *Handler, authMiddleware, managerOrAdmin gin.HandlerFunc) {
	group := api.Group("/parking-slots")
	{
		group.POST("", authMiddleware, managerOrAdmin, handler.Create)
		group.GET("/:id", authMiddleware, handler.FindByID)
		group.PATCH("/admin/:id", authMiddleware, managerOrAdmin, handler.AdminUpdateStatus)
		group.POST("/sensor", handler.SensorUpdateStatus)
		group.PATCH("/:id/device", authMiddleware, managerOrAdmin, handler.ChangeDevice)
	}
}
