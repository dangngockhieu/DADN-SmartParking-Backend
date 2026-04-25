package rfid_card

import "github.com/gin-gonic/gin"

func RegisterRoutes(api *gin.RouterGroup, handler *Handler) {
	group := api.Group("/rfid-cards")
	{
		group.POST("", handler.Create)
		group.PATCH("/:id", handler.Update)
	}
}
