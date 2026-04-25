package slot_history

import "github.com/gin-gonic/gin"

func RegisterRoutes(api *gin.RouterGroup, handler *Handler, authMiddleware, managerOrAdmin gin.HandlerFunc) {
	group := api.Group("/slot-histories")
	group.Use(authMiddleware)
	{
		group.GET("/:slotId", managerOrAdmin, handler.GetBySlotID)
	}
}
