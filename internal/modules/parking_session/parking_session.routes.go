package parking_session

import "github.com/gin-gonic/gin"

func RegisterRoutes(api *gin.RouterGroup, handler *Handler) {
	group := api.Group("/parking-sessions")
	{
		group.GET("", handler.FindAll)
		group.GET("/:id", handler.FindByID)
		group.DELETE("/:id", handler.ForceClose)
		group.DELETE("/card/:uid", handler.CloseByCard)
		group.DELETE("/purge/:id", handler.HardDelete)
	}
}
