package main

import (
	"log"
	"net/http"
	"os"

	"backend/configs"
	_ "backend/docs"
	"backend/internal/auth"
	authmail "backend/internal/auth/mail"
	"backend/internal/auth/token"
	"backend/internal/modules/gate"
	"backend/internal/modules/iot_device"
	"backend/internal/modules/iot_gateway"
	"backend/internal/modules/parking_lot"
	"backend/internal/modules/parking_session"
	"backend/internal/modules/parking_slot"
	"backend/internal/modules/rfid_card"
	"backend/internal/modules/slot_history"
	"backend/internal/modules/user"
	"backend/internal/realtime/parking"
	"backend/pkg/database"
	"backend/pkg/middleware"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title Smart Parking Backend API
// @version 1.0
// @description Backend API cho hệ thống smart parking
// @host localhost:8080
// @BasePath /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	cfg := configs.LoadConfig()

	db := database.NewMySQL(cfg)
	redisClient := database.NewRedis(cfg)
	defer redisClient.Close()

	tokenService := token.NewService(cfg)
	mailService := authmail.NewService(cfg)

	authMiddleware := middleware.Auth(tokenService, redisClient)
	adminOnly := middleware.RequireRoles(user.RoleAdmin)
	managerOrAdmin := middleware.RequireRoles(user.RoleManager, user.RoleAdmin)

	// realtime hub
	parkingHub := parking.NewHub()

	// modules
	authModule := auth.NewModule(db, redisClient, tokenService, mailService)
	iotDeviceModule := iot_device.NewModule(db)
	slotHistoryModule := slot_history.NewModule(db)
	parkingLotModule := parking_lot.NewModule(db)
	gateModule := gate.NewModule(db)
	userModule := user.NewModule(db)
	rfidCardModule := rfid_card.NewModule(db)
	parkingSessionModule := parking_session.NewModule(db)
	iotGatewayModule := iot_gateway.NewModule(gateModule.Service, rfidCardModule.Service, parkingSessionModule.Service)
	parkingSlotModule := parking_slot.NewModule(db, parkingHub)

	// webtransport server chạy riêng
	if _, certErr := os.Stat("cert.pem"); certErr == nil {
		if _, keyErr := os.Stat("key.pem"); keyErr == nil {
			go func() {
				wtServer := parking.NewServer(parkingHub, "cert.pem", "key.pem")
				if err := wtServer.Run(":8443"); err != nil {
					log.Printf("webtransport server stopped: %v", err)
				}
			}()
		} else {
			log.Printf("skip webtransport: key file not found: %v", keyErr)
		}
	} else {
		log.Printf("skip webtransport: cert file not found: %v", certErr)
	}

	r := gin.New()
	if err := r.SetTrustedProxies([]string{"127.0.0.1", "::1"}); err != nil {
		log.Fatalf("set trusted proxies failed: %v", err)
	}

	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(middleware.CORS(cfg))
	r.Use(middleware.ErrorHandler())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "up",
			"message": "Service is running perfectly",
		})
	})

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	api := r.Group("/api/v1")

	auth.RegisterRoutes(api, authModule.Handler, authMiddleware)
	iot_device.RegisterRoutes(api, iotDeviceModule.Handler, authMiddleware, managerOrAdmin)
	slot_history.RegisterRoutes(api, slotHistoryModule.Handler, authMiddleware, managerOrAdmin)
	parking_lot.RegisterRoutes(api, parkingLotModule.Handler, authMiddleware, managerOrAdmin)
	parking_slot.RegisterRoutes(api, parkingSlotModule.Handler, authMiddleware, managerOrAdmin)
	iot_gateway.RegisterRoutes(api, iotGatewayModule.Handler)
	gate.RegisterRoutes(api, gateModule.Handler, authMiddleware, managerOrAdmin)
	user.RegisterRoutes(api, userModule.Handler, authMiddleware, adminOnly)
	rfid_card.RegisterRoutes(api, rfidCardModule.Handler)
	parking_session.RegisterRoutes(api, parkingSessionModule.Handler)

	_ = r.Run(":" + cfg.AppPort)
}
