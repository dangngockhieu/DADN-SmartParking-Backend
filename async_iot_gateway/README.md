# Async IoT Gateway

Service này nhận dữ liệu IoT theo mô hình async:

```text
ESP32 sensor batch
  -> FastAPI Async Gateway
  -> Redis Streams
  -> Async Sensor Worker
  -> MySQL
```

Luồng gate vẫn là request-response vì cần trả ngay `open_barrier` hoặc `reject`:

```text
Camera/RFID
  -> FastAPI Async Gateway
  -> Redis PlateCache + MySQL
  -> response cho ESP32 gate
```

## Cài đặt

Chạy từ root project:

```bash
cd async_iot_gateway
python -m venv .venv
.venv\Scripts\activate
pip install -r requirements.txt
```

Service đọc chung biến môi trường với Go backend trong `.env.local`:

```env
DB_HOST=127.0.0.1
DB_PORT=3307
DB_USER=...
DB_PASS=...
DB_NAME=...
REDIS_ADDR=127.0.0.1:6379
REDIS_PASSWORD=
IOT_GATEWAY_PORT=8090
```

## Chạy API gateway

```bash
uvicorn app.main:app --host 0.0.0.0 --port 8090 --reload
```

## Chạy sensor worker

Mở terminal khác:

```bash
python -m app.worker
```

## Endpoints

```http
POST /api/v1/iot/sensor/batch
POST /api/v1/iot/camera
POST /api/v1/iot/rfid
GET  /health
```

## Sensor batch payload

```json
{
  "device_mac": "SENSOR_A_001",
  "lot_id": 1,
  "sequence": 120,
  "sent_at": "2026-05-13T10:30:00+07:00",
  "events": [
    {
      "port": 1,
      "is_occupied": true,
      "changed_at": "2026-05-13T10:29:58+07:00"
    }
  ]
}
```

FastAPI chỉ ghi batch vào Redis Stream và trả ACK nhanh. Worker mới là nơi gom batch, bỏ trùng, so sánh trạng thái và cập nhật MySQL.
