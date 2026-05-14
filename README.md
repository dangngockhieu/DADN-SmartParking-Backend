# Smart Parking Backend

Backend API cho hệ thống bãi xe thông minh, thiết kế theo kiến trúc High-Performance phân tán, sử dụng Go (Gin) + Python (FastAPI) + MySQL + Redis.

## 1) Tính năng chính

- Xác thực người dùng bằng JWT + refresh token.
- Quản lý bãi xe, cổng (gate), thiết bị IoT, thẻ RFID, vị trí đỗ xe.
- **Kiến trúc IoT & Realtime chịu tải cao (10k+ concurrent users)**:
  - Ingestion siêu tốc qua FastAPI & Redis Streams.
  - Phân tích & Gộp dữ liệu (Coalescing/Batching) qua Python Worker.
  - Phân phối realtime qua Go Server với Sharded Hub & WebTransport (HTTP/3).
- Nhận dữ liệu realtime từ thiết bị:
  - Camera nhận diện biển số.
  - RFID scan tại cổng vào/ra (tích hợp Transaction thanh toán).
  - Sensor cập nhật trạng thái ô đỗ (hỗ trợ đẩy theo Batch).
- Dashboard thống kê lưu lượng xe, tỉ lệ lấp đầy theo bãi.
- Ví điện tử + nạp tiền qua PayOS (deposit, webhook, cancel-return).
- Tài liệu API qua Swagger.

## 2) Tech Stack

- **Core API & Realtime Broadcast**: Go 1.25 (Gin, WebTransport, Sharded Hub)
- **IoT Ingestion Gateway**: Python 3.10+ (FastAPI, Uvicorn)
- **Async Worker**: Python (asyncio, batch processing)
- **Database**: GORM + MySQL 8+
- **Message Broker & Cache**: Redis 6+ (Streams, Pub/Sub, Hash)
- **Docs**: Swagger (`swaggo/gin-swagger`)

## 3) Cấu trúc thư mục

```text
backend
├── async_iot_gateway/          # (Mới) Python FastAPI Gateway & Async Worker
│   ├── app/
│   │   ├── main.py             # Entry point FastAPI (:8090)
│   │   ├── worker.py           # Background Worker xử lý batch sensor
│   │   ├── gate_service.py     # Logic xử lý camera & rfid (Lock, cache)
│   │   └── ...
│   └── requirements.txt
│
├── cmd
│   ├── api/main.go             # Entry point Go API server (:8080)
│   └── main/seed.go            # Chạy seed dữ liệu mẫu
│
├── configs/                    # Go Config loader
├── docs/                       # Docs swagger
│
├── internal
│   ├── auth/                   # Xử lý đăng nhập, JWT, Mail
│   ├── common/errors/          # Format response error
│   │
│   ├── modules/                # Các module nghiệp vụ (Go)
│   │   ├── dashboard/          # Thống kê lưu lượng
│   │   ├── gate/               # Quản lý thiết lập cổng
│   │   ├── parking_session/    # Quản lý phiên đỗ xe, tính tiền ra/vào
│   │   ├── parking_slot/       # Quản lý ô đỗ
│   │   ├── wallet/             # Ví điện tử & Tích hợp nạp tiền PayOS
│   │   └── ...                 # user, rfid_card, iot_device...
│   │
│   └── realtime/parking/       # Go WebTransport Server & Sharded Hub
│
├── migrations/                 # Các file SQL tạo bảng DB
├── pkg
│   ├── database/               # Khởi tạo connect MySQL, Redis (Go)
│   └── middleware/             # Go Middleware (Auth, CORS...)
│
├── templates/                  # Template HTML gửi email
└── ...
```

## 4) Kiến trúc Hệ Thống (High-load Realtime Architecture)

Hệ thống được chia làm 2 mảnh ghép chính hoạt động song song và giao tiếp qua Redis:

1. **FastAPI IoT Gateway (`:8090`)**: Chuyên tiếp nhận dữ liệu từ phần cứng (ESP32, Camera, RFID) cực nhanh, đẩy vào Redis Stream và phản hồi ngay lập tức (`202 Accepted`). Không kết nối MySQL trực tiếp ở luồng sensor để tránh thắt cổ chai.
2. **Go API Server (`:8080` & `:4433`)**: Xử lý toàn bộ Business Logic (User, Wallet, CRUD) và chịu trách nhiệm làm **Realtime Hub**. Nhận tín hiệu thay đổi từ Redis Pub/Sub, khoanh vùng user theo từng bãi xe (Lot Room) và Fan-out qua WebTransport tới Frontend bằng Worker Pool.
3. **Python Async Worker**: Chạy ngầm, đọc dữ liệu thô từ Redis Stream, gộp/khử nhiễu (Deduplicate) trong cửa sổ thời gian, cập nhật batch vào MySQL và bắn tín hiệu báo Frontend.

## 5) Chuẩn bị môi trường

### 5.1 Yêu cầu

- Go (khuyến nghị khớp với `go.mod`)
- Python 3.10+ (để chạy FastAPI Gateway)
- MySQL 8+
- Redis 6+

### 5.2 Dùng Docker cho MySQL & Redis

Repo đã có `docker-compose.yaml` cho MySQL:

```bash
docker compose up -d
```

Mặc định compose map MySQL ra cổng `3307`.

Mở terminal và chạy câu lệnh sau để tạo db redis:

```bash
docker run --name parking -p 6379:6379 -d redis
```

Sẽ tạo ra db redis chạy trên cổng `6379`.

## 6) Cấu hình môi trường

Project đọc `.env.local` khi `APP_ENV=local`.
1. Tạo file `.env.local` từ `.env.example`.
2. Điền giá trị phù hợp. Chú ý cấu hình Redis và MySQL.

Ví dụ tham khảo trong file `.env.example` đã có sẵn. Cần đảm bảo `REDIS_ADDR` và thông tin MySQL chính xác cho cả Go và FastAPI cùng truy cập.

## 7) Khởi tạo database schema & Seed dữ liệu

Hiện repo dùng SQL migration files trong `migrations/`.

Có thể áp dụng thủ công file up bằng MySQL client:
```bash
mysql -h 127.0.0.1 -P 3307 -u admin -p parking < migrations/000001_init.up.sql
```

Chạy seed dữ liệu mẫu:
```bash
go run ./cmd/main/seed.go
```

Dữ liệu mẫu gồm các tài khoản admin, guest, và các IoT Devices/Gates mẫu.

## 8) Hướng dẫn Chạy Server (Yêu cầu chạy cả 2 tiến trình)

Để toàn bộ hệ thống IoT và Realtime hoạt động, bạn cần mở **2 Terminal** để chạy song song Go Server và Python FastAPI.

### Terminal 1: Chạy Go API Server & WebTransport
Đảm nhiệm giao diện Admin, WebTransport cho Frontend.
```bash
cd backend
go run ./cmd/api
```
- API chạy ở cổng `:8080`
- WebTransport chạy ở cổng `:4433` (yêu cầu file cert.pem và key.pem)
- Swagger UI: `http://localhost:8080/swagger/index.html`

### Terminal 2: Chạy FastAPI IoT Gateway
Đảm nhiệm tiếp nhận dữ liệu từ ESP32, RFID, Camera.
```bash
cd backend/async_iot_gateway

# 1. Kích hoạt môi trường ảo (hoặc tạo mới: python -m venv .venv)
# Windows: ..\.venv\Scripts\Activate.ps1
# Mac/Linux: source ../.venv/bin/activate
pip install -r requirements.txt

# 2. Chạy server FastAPI
uvicorn app.main:app --host 0.0.0.0 --port 8090
```
- Gateway chạy ở cổng `:8090`

### Terminal 3 (Tùy chọn): Chạy Python Sensor Worker
Xử lý dữ liệu từ Redis Stream vào MySQL.
```bash
cd backend/async_iot_gateway
python -m app.worker
```

## 9) Luồng IoT API (Gọi vào FastAPI `:8090`)

### 9.1 Camera -> Cache biển số
```http
POST http://localhost:8090/api/v1/iot/camera
Content-Type: application/json

{
  "gate_id": 1,
  "plate_number": "51A-123.45"
}
```

### 9.2 RFID scan -> Giao dịch ra/vào
Phải gọi Camera trước để có biển số trong cache, sau đó quẹt thẻ.
```http
POST http://localhost:8090/api/v1/iot/rfid
Content-Type: application/json

{
  "gate_id": 1,
  "mac_address": "GATE_IN_A_001",
  "rfid_uid": "USER001"
}
```

### 9.3 ESP32 Sensor -> Batch cập nhật trạng thái ô đỗ
Bắn dữ liệu hàng loạt không lo tắc nghẽn (trả về HTTP 202 ngay lập tức).
```http
POST http://localhost:8090/api/v1/iot/sensor/batch
Content-Type: application/json

{
  "device_mac": "SENSOR_A_001",
  "sequence": 12345,
  "events": [
    {
      "port": 1,
      "is_occupied": true
    },
    {
      "port": 2,
      "is_occupied": false
    }
  ]
}
```

## 10) Tùy chỉnh Độ Trễ Realtime (Latency Tuning)
Mặc định hệ thống gom batch mỗi `1000ms` (1 giây) để bảo vệ Database khi quá tải và khử nhiễu cảm biến. Nếu muốn hệ thống chớp trạng thái ngay lập tức (độ trễ < 50ms) cho mục đích biểu diễn realtime siêu mượt, bạn có thể chỉnh biến sau trong `.env` (hoặc `.env.local`):
```env
IOT_WORKER_BATCH_WINDOW_MS=50
```
Sau đó khởi động lại Worker.

## 11) Route groups chính (Bên Go `:8080`)

- `/api/v1/auth`, `/api/v1/users`
- `/api/v1/parking-lots`, `/api/v1/parking-slots`, `/api/v1/parking-sessions`
- `/api/v1/gates`, `/api/v1/rfid-cards`, `/api/v1/iot-devices`
- `/api/v1/dashboard`, `/api/v1/wallets`
