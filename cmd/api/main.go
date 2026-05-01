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

// getCertPaths trả về đường dẫn certificate
// Dev: dùng mặc định cert.pem, key.pem
// Production: đọc từ environment variables
func getCertPaths() (certPath, keyPath string) {
	certPath = os.Getenv("TLS_CERT")
	keyPath = os.Getenv("TLS_KEY")

	// Fallback cho dev environment
	if certPath == "" {
		certPath = "cert.pem"
	}
	if keyPath == "" {
		keyPath = "key.pem"
	}

	return certPath, keyPath
}

func main() {
	cfg := configs.LoadConfig()

	db := database.NewMySQL(cfg)
	redisClient := database.NewRedis(cfg)
	defer redisClient.Close()

	tokenService := token.NewService(cfg)
	mailService := authmail.NewService(cfg)

	authMiddleware := middleware.Auth(tokenService, redisClient)
	adminOnly := middleware.RequireRoles(user.RoleAdmin)

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
	parkingSlotModule := parking_slot.NewModule(db, parkingHub)
	// rfidCardModule := rfid_card.NewModule(db, userModule.Service)
	parkingSessionModule := parking_session.NewModule(db)
	iotGatewayModule := iot_gateway.NewModule(gateModule.Service, rfidCardModule.Service, parkingSessionModule.Service, parkingSlotModule.Service)

	// Khởi động WebTransport server
	certPath, keyPath := getCertPaths()
	go func() {
		if _, err := os.Stat(certPath); err != nil {
			log.Printf("WebTransport disabled: cert file not found at %s", certPath)
			return
		}
		if _, err := os.Stat(keyPath); err != nil {
			log.Printf("WebTransport disabled: key file not found at %s", keyPath)
			return
		}

		wtServer := parking.NewServer(parkingHub, certPath, keyPath)
		log.Printf("Starting WebTransport server on :8443 (cert: %s, key: %s)", certPath, keyPath)
		if err := wtServer.Run(":8443"); err != nil {
			log.Printf("WebTransport server stopped: %v", err)
		}
	}()

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
	iot_device.RegisterRoutes(api, iotDeviceModule.Handler, authMiddleware, adminOnly)
	slot_history.RegisterRoutes(api, slotHistoryModule.Handler, authMiddleware, adminOnly)
	parking_lot.RegisterRoutes(api, parkingLotModule.Handler, authMiddleware, adminOnly)
	parking_slot.RegisterRoutes(api, parkingSlotModule.Handler, authMiddleware, adminOnly)
	iot_gateway.RegisterRoutes(api, iotGatewayModule.Handler)
	gate.RegisterRoutes(api, gateModule.Handler, authMiddleware, adminOnly)
	user.RegisterRoutes(api, userModule.Handler, authMiddleware, adminOnly)
	rfid_card.RegisterRoutes(api, rfidCardModule.Handler, authMiddleware, adminOnly)
	parking_session.RegisterRoutes(api, parkingSessionModule.Handler)

	_ = r.Run(":" + cfg.AppPort)
}
