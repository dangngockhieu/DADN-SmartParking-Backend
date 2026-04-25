package user

import "github.com/gin-gonic/gin"

func RegisterRoutes(api *gin.RouterGroup, handler *Handler, authMiddleware, adminOnly gin.HandlerFunc) {
	group := api.Group("/users")
	group.Use(authMiddleware)
	{
		group.GET("", adminOnly, handler.FindWithPagination)
		group.POST("", adminOnly, handler.CreateByAdmin)
		group.PATCH("/change-password", handler.ChangePassword)
		group.PATCH("/change-role/:id", adminOnly, handler.ChangeRole)
	}
}
