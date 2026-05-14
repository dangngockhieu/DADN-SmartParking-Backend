package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"backend/configs"
	_ "backend/docs"
	"backend/internal/auth"
	authmail "backend/internal/auth/mail"
	"backend/internal/auth/token"
	"backend/internal/modules/dashboard"
	"backend/internal/modules/gate"
	"backend/internal/modules/iot_device"
	"backend/internal/modules/parking_lot"
	"backend/internal/modules/parking_session"
	"backend/internal/modules/parking_slot"
	"backend/internal/modules/rfid_card"
	"backend/internal/modules/user"

	"backend/internal/modules/wallet"
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

// getCertPaths trả về đường dẫn certificate.
// Dev: dùng mặc định cert.pem, key.pem.
// Production: đọc từ environment variables TLS_CERT, TLS_KEY.
func getCertPaths() (certPath, keyPath string) {
	// Thiết lập kết nối TLS cho WebTransport
	certPath = os.Getenv("TLS_CERT")
	keyPath = os.Getenv("TLS_KEY")

	if certPath == "" {
		certPath = "cert.pem"
	}
	if keyPath == "" {
		keyPath = "key.pem"
	}

	return certPath, keyPath
}

// slotChangeItem là struct cho 1 event thay đổi slot (dùng trong bridge).
type slotChangeItem struct {
	Changed   bool   `json:"changed"`
	ID        uint64 `json:"id"`
	LotID     uint64 `json:"lot_id"`
	Name      string `json:"name"`
	OldStatus string `json:"old_status"`
	NewStatus string `json:"new_status"`
	Message   string `json:"message"`
}

// startRealtimeBridge subscribes Redis pub/sub channel "iot:parking:slot_status_changed"
// và forward events từ Python Worker → ParkingHub → WebTransport → Frontend.
//
// Tối ưu: group events theo lotID → gửi batch 1 lần per lot (giảm lock + send).
// Luồng: Python Sensor Worker → Redis PubSub → Go Bridge → ParkingHub.BroadcastBatchToLot → Frontend
func startRealtimeBridge(redisClient *database.RedisClient, hub *parking.Hub) {
	go func() {
		ctx := context.Background()
		pubsub := redisClient.Client.Subscribe(ctx, "iot:parking:slot_status_changed")
		defer pubsub.Close()

		log.Println("[BRIDGE] subscribed to Redis channel: iot:parking:slot_status_changed")

		ch := pubsub.Channel()
		for msg := range ch {
			var envelope struct {
				Event string `json:"event"`
				Data  []struct {
					ID        uint64 `json:"id"`
					LotID     uint64 `json:"lot_id"`
					Name      string `json:"name"`
					OldStatus string `json:"old_status"`
					NewStatus string `json:"new_status"`
				} `json:"data"`
			}

			if err := json.Unmarshal([]byte(msg.Payload), &envelope); err != nil {
				log.Printf("[BRIDGE] invalid JSON from Redis: %v", err)
				continue
			}

			if envelope.Event != "SLOT_STATUS_CHANGE_BATCH" || len(envelope.Data) == 0 {
				continue
			}

			// Group events theo lotID → giảm số lần broadcast từ N events → M lots
			byLot := make(map[uint64][]slotChangeItem)
			for _, item := range envelope.Data {
				byLot[item.LotID] = append(byLot[item.LotID], slotChangeItem{
					Changed:   true,
					ID:        item.ID,
					LotID:     item.LotID,
					Name:      item.Name,
					OldStatus: item.OldStatus,
					NewStatus: item.NewStatus,
					Message:   "Cập nhật từ sensor",
				})
			}

			// Gửi batch cho mỗi lot (1 lần lock + 1 lần send per client per lot)
			for lotID, events := range byLot {
				hub.BroadcastBatchToLot(lotID, "SLOT_STATUS_CHANGE_BATCH", events)
			}

			log.Printf("[BRIDGE] forwarded %d slot changes across %d lots (hub: %d clients, %d rooms)",
				len(envelope.Data), len(byLot), hub.ClientCount(), hub.RoomCount())
		}

		log.Println("[BRIDGE] Redis subscription ended")
	}()
}

func main() {
	// Load giá trị từ file .env
	cfg := configs.LoadConfig()

	// Kết nối database và Redis
	db := database.NewMySQL(cfg)
	redisClient := database.NewRedis(cfg)
	defer redisClient.Close()

	// Khởi tạo services, modules, middleware
	tokenService := token.NewService(cfg)
	mailService := authmail.NewService(cfg)

	authMiddleware := middleware.Auth(tokenService, redisClient)
	adminOnly := middleware.RequireRoles(user.RoleAdmin)

	// Realtime hub dùng chung cho WebTransport và parking_slot service
	parkingHub := parking.NewHub()

	// Modules
	authModule := auth.NewModule(db, redisClient, tokenService, mailService)
	iotDeviceModule := iot_device.NewModule(db)
	parkingLotModule := parking_lot.NewModule(db)
	gateModule := gate.NewModule(db)
	userModule := user.NewModule(db)
	rfidCardModule := rfid_card.NewModule(db)
	dashboardModule := dashboard.NewModule(db)
	walletModule := wallet.NewModule(db, cfg)

	// Quan trọng: inject parkingHub vào parking_slot module
	parkingSlotModule := parking_slot.NewModule(db, parkingHub)

	parkingSessionModule := parking_session.NewModule(db)

	// Khởi động WebTransport server
	certPath, keyPath := getCertPaths()
	go func() {
		if _, err := os.Stat(certPath); err != nil {
			log.Printf("[WT] disabled: cert file not found at %s, err=%v", certPath, err)
			return
		}

		if _, err := os.Stat(keyPath); err != nil {
			log.Printf("[WT] disabled: key file not found at %s, err=%v", keyPath, err)
			return
		}

		wtServer := parking.NewServer(parkingHub, certPath, keyPath, cfg)

		log.Printf("[WT] starting WebTransport server on :4433 (cert: %s, key: %s)", certPath, keyPath)

		if err := wtServer.Run(":4433"); err != nil {
			log.Printf("[WT] server stopped: %v", err)
		}
	}()

	// Khởi động Redis Bridge: Python Worker → Redis PubSub → ParkingHub → WebTransport → Frontend
	startRealtimeBridge(redisClient, parkingHub)

	r := gin.New()

	if err := r.SetTrustedProxies([]string{"127.0.0.1", "::1"}); err != nil {
		log.Fatalf("[APP] set trusted proxies failed: %v", err)
	}

	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(middleware.CORS(cfg))
	r.Use(middleware.ErrorHandler())

	// Endpoint health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "up",
			"message": "Service is running perfectly",
		})
	})

	// Swagger UI available at /swagger/index.html
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Prefix chung cho API versioning
	api := r.Group("/api/v1")

	// Đăng ký routes cho từng module
	auth.RegisterRoutes(api, authModule.Handler, authMiddleware)
	iot_device.RegisterRoutes(api, iotDeviceModule.Handler, authMiddleware, adminOnly)
	parking_lot.RegisterRoutes(api, parkingLotModule.Handler, authMiddleware, adminOnly)
	parking_slot.RegisterRoutes(api, parkingSlotModule.Handler, authMiddleware, adminOnly)
	gate.RegisterRoutes(api, gateModule.Handler, authMiddleware, adminOnly)
	user.RegisterRoutes(api, userModule.Handler, authMiddleware, adminOnly)
	rfid_card.RegisterRoutes(api, rfidCardModule.Handler, authMiddleware, adminOnly)
	parking_session.RegisterRoutes(api, parkingSessionModule.Handler, authMiddleware, adminOnly)
	dashboard.RegisterRoutes(api, dashboardModule.Handler, authMiddleware, adminOnly)
	wallet.RegisterRoutes(api, walletModule.Handler, authMiddleware, adminOnly)

	log.Printf("[APP] starting Gin server on :%s", cfg.AppPort)

	if err := r.Run(":" + cfg.AppPort); err != nil {
		log.Fatalf("[APP] Gin server stopped: %v", err)
	}
}
