# Smart Parking Scaling Architecture Options

> Cập nhật 2026-05-13: Phương án MQTT không còn là hướng triển khai hiện tại. Kiến trúc mới đã được implement trong `async_iot_gateway/`: ESP32 gửi HTTP batch lên FastAPI Async Gateway, Gateway ghi Redis Streams, Async Worker gom batch và ghi MySQL. Luồng gate vẫn request-response qua FastAPI để trả ngay `open_barrier` hoặc `reject`.

Tài liệu này mô tả 2 phương án thay đổi kiến trúc giao tiếp giữa embedded và backend để hệ thống smart parking có thể xử lý số lượng lớn sensor cập nhật đồng thời.

Hiện tại backend nhận dữ liệu sensor qua HTTP request từng sự kiện. Mỗi lần sensor thay đổi, ESP32 gọi API, backend query MySQL để tìm slot, update trạng thái, broadcast realtime rồi trả response. Cách này chạy tốt với demo nhỏ, nhưng khi có hàng nghìn hoặc hàng chục nghìn sensor gửi cùng lúc, MySQL và HTTP handler sẽ bị áp lực lớn vì mỗi sensor update có thể tạo ra nhiều thao tác đồng bộ.

Mục tiêu của 2 option dưới đây là tách luồng sensor ra khỏi luồng xử lý nghiệp vụ cần phản hồi ngay. Sensor status là telemetry stream, không nên bắt request sensor chờ ghi database. Trong khi đó RFID/gate là command decision flow, vẫn cần phản hồi ngay để quyết định mở barrier hoặc từ chối.

## Option 1: HTTP Batch + Redis Worker

Option 1 giữ HTTP làm giao thức giao tiếp giữa embedded và backend, nhưng thay đổi cách gửi và xử lý dữ liệu sensor. Thay vì gửi từng sensor update riêng lẻ, ESP32 hoặc board gateway gửi một gói batch gồm nhiều trạng thái slot trong cùng một request.

### Các khối chính

**ESP32 Sensors**

ESP32 đọc dữ liệu từ các cảm biến SR04 gắn với từng ô đỗ. Mỗi sensor cho biết ô đỗ đang có xe hay trống. ESP32 không gửi request ngay sau mỗi lần đo raw distance, mà phải có lớp xử lý cục bộ:

- Đọc khoảng cách theo chu kỳ ngắn, ví dụ 200-500 ms.
- Lọc nhiễu bằng debounce, ví dụ chỉ xác nhận đổi trạng thái nếu cùng kết quả xuất hiện liên tiếp 3-5 lần.
- Chỉ đưa event vào batch khi trạng thái logic thay đổi từ `AVAILABLE` sang `OCCUPIED` hoặc ngược lại.
- Nếu không có thay đổi, không gửi gì.

**HTTP Batch API**

Backend cung cấp endpoint nhận batch, ví dụ:

```http
POST /api/v1/parking-slots/sensor/batch
```

Endpoint này chỉ làm việc nhẹ:

- Parse JSON.
- Validate `device_mac`, `port`, `is_occupied`, `timestamp`, `sequence`.
- Ghi trạng thái mới nhất vào Redis.
- Đưa event vào Redis queue hoặc Redis Stream.
- Trả response nhanh, ví dụ `202 Accepted`.

Endpoint này không nên query rồi update MySQL từng slot ngay trong request.

**Redis Latest State + Queue**

Redis đóng 2 vai trò:

- Latest state: lưu trạng thái mới nhất của từng slot theo key như `slot:{device_mac}:{port}`.
- Queue/stream: lưu event để worker xử lý sau, ví dụ dùng Redis Stream `slot_status_events`.

Redis nằm giữa HTTP handler và worker để hấp thụ burst traffic. Khi 10.000 sensor gửi cùng lúc, request chỉ cần ghi Redis rất nhanh thay vì đồng loạt đập vào MySQL.

**Sensor Worker**

Worker chạy nền trong backend hoặc service riêng. Worker đọc event từ Redis queue/stream, gom nhiều event trong một cửa sổ thời gian ngắn, ví dụ 1-5 giây, sau đó xử lý theo batch.

Worker cần làm:

- Gộp event trùng theo `device_mac + port`.
- Chỉ giữ trạng thái cuối cùng nếu một slot đổi nhiều lần trong cùng batch window.
- So sánh với latest state hoặc cache nội bộ để bỏ update không đổi.
- Batch update MySQL bằng một transaction hoặc nhiều câu update tối ưu.
- Publish event realtime cho frontend sau khi trạng thái đã được chấp nhận.

**MySQL**

MySQL vẫn là nguồn dữ liệu bền vững cho `parking_slots`, `parking_sessions`, `users`, `rfid_cards`, wallet và lịch sử. Tuy nhiên MySQL không còn nhận 10.000 update trực tiếp trong cùng một thời điểm từ HTTP handler.

Với slot status, MySQL nên được update theo batch. Nếu cần phân tích lịch sử cảm biến, có thể thêm bảng event/history riêng, nhưng trạng thái hiện tại của slot chỉ cần lưu bản mới nhất.

**ParkingHub / WebTransport**

Sau khi worker xử lý xong batch, backend push event realtime sang frontend. Frontend không cần biết từng sensor raw event, chỉ cần biết slot nào đổi trạng thái.

**Camera / RFID Gate**

Luồng gate vẫn giữ HTTP đồng bộ vì gate cần phản hồi ngay:

- Camera gửi biển số lên backend.
- Backend lưu plate tạm vào Redis PlateCache theo `gate_id`.
- RFID gửi `gate_id`, `mac_address`, `rfid_uid`.
- Backend kiểm tra gate, card, plate, session, ví.
- Backend trả `open_barrier` hoặc `reject`.

PlateCache nên chuyển từ RAM local sang Redis TTL để backend có thể scale nhiều instance.

### Sensor batch gửi ra sao

Payload đề xuất:

```json
{
  "device_mac": "SENSOR_A_001",
  "lot_id": 1,
  "sequence": 1024,
  "sent_at": "2026-05-12T10:30:00+07:00",
  "events": [
    {
      "port": 1,
      "is_occupied": true,
      "changed_at": "2026-05-12T10:29:58+07:00"
    },
    {
      "port": 3,
      "is_occupied": false,
      "changed_at": "2026-05-12T10:29:59+07:00"
    }
  ]
}
```

`sequence` giúp backend phát hiện duplicate hoặc out-of-order batch. `changed_at` là thời điểm sensor xác nhận trạng thái đổi, còn `sent_at` là thời điểm ESP32 gửi batch.

### Khi nào sensor gửi lên

ESP32 nên gửi batch theo 3 điều kiện:

- Gửi ngay khi số event trong batch đạt ngưỡng, ví dụ 8-32 event.
- Gửi định kỳ nếu batch đang có dữ liệu, ví dụ mỗi 1-5 giây.
- Gửi heartbeat thưa hơn nếu không có thay đổi, ví dụ mỗi 30-60 giây, chỉ để báo thiết bị còn sống.

Không gửi mỗi lần đo khoảng cách. Chỉ gửi khi trạng thái logic đã đổi sau debounce.

### Dữ liệu đi như thế nào

Luồng sensor:

```text
ESP32 Sensors
  -> HTTP Batch API
  -> Redis Latest State + Queue
  -> Sensor Worker
  -> MySQL parking_slots
  -> ParkingHub / WebTransport
  -> Frontend
```

Luồng gate:

```text
Camera / RFID Gate
  -> HTTP IoT API
  -> Redis PlateCache
  -> MySQL parking_sessions / wallet
  -> HTTP response open_barrier / reject
  -> Gate actuator
```

### Vì sao xử lý được concurrency lớn

Option 1 xử lý concurrency tốt hơn kiến trúc hiện tại vì request sensor không còn giữ kết nối để chờ MySQL. HTTP handler chỉ làm thao tác nhanh với Redis và trả response. Redis có khả năng chịu tải ghi cao hơn MySQL cho workload key-value/queue ngắn.

Khi 10.000 sensor cùng cập nhật:

- HTTP layer có thể scale ngang nhiều instance sau load balancer.
- Redis hấp thụ burst traffic.
- Worker kiểm soát tốc độ ghi MySQL bằng batch size và batch interval.
- MySQL nhận ít câu lệnh hơn vì event trùng được gộp.
- Frontend nhận event đã lọc thay vì nhận raw event liên tục.

Điểm quan trọng là hệ thống chuyển từ synchronous write path sang asynchronous ingestion path. Sensor không phụ thuộc trực tiếp vào tốc độ ghi database.

### Ưu điểm và hạn chế

Ưu điểm:

- Ít thay đổi firmware hơn so với MQTT.
- Vẫn dùng HTTP quen thuộc.
- Dễ tích hợp vào backend Go hiện tại.
- Giảm áp lực MySQL đáng kể.

Hạn chế:

- HTTP vẫn có overhead lớn hơn MQTT cho số lượng thiết bị rất lớn.
- Mỗi batch vẫn là một HTTP request độc lập.
- Cần tự xử lý retry/idempotency nếu request lỗi.

## Option 2: MQTT Broker + Batch Worker

Option 2 chuyển luồng sensor sang MQTT. Đây là hướng phù hợp hơn cho IoT thực tế. Theo yêu cầu, sensor trong Option 2 cũng gửi theo batch, nhưng batch được publish qua MQTT thay vì HTTP.

Gate/RFID có thể vẫn giữ HTTP để quyết định mở barrier ngay. Sensor status và gate decision nên tách riêng vì bản chất khác nhau: sensor là telemetry async, gate là nghiệp vụ cần phản hồi đồng bộ.

### Các khối chính

**ESP32 Sensors**

ESP32 đọc SR04, debounce và gom batch giống Option 1. Khác biệt là thay vì gọi HTTP API, ESP32 publish batch lên MQTT topic.

ESP32 cần có:

- Task đọc sensor.
- Queue nội bộ trong RAM để chứa event đã debounce.
- MQTT publish task chuyên gửi batch.
- Cơ chế reconnect MQTT.
- Cơ chế giữ batch chưa gửi nếu mất mạng trong thời gian ngắn, tùy giới hạn RAM/flash.

**MQTT Publish**

ESP32 publish batch lên topic có cấu trúc rõ ràng, ví dụ:

```text
parking/{lot_id}/sensor/{device_mac}/batch
```

Ví dụ:

```text
parking/1/sensor/SENSOR_A_001/batch
```

Payload vẫn là JSON batch, tương tự Option 1.

**MQTT Broker**

Broker như Mosquitto hoặc EMQX đứng giữa thiết bị và backend. Broker chịu trách nhiệm:

- Quản lý hàng nghìn kết nối thiết bị.
- Nhận publish từ ESP32.
- Chuyển message đến backend subscriber.
- Hỗ trợ QoS, retained message, session, authentication tùy cấu hình.

Với hệ thống lớn, EMQX phù hợp hơn Mosquitto vì hỗ trợ clustering và quản trị tốt hơn. Với đồ án hoặc demo mở rộng, Mosquitto là đủ để trình bày.

**Sensor Worker**

Sensor Worker subscribe các topic sensor batch:

```text
parking/+/sensor/+/batch
```

Worker xử lý:

- Parse payload.
- Validate `lot_id`, `device_mac`, `sequence`, danh sách events.
- Ghi latest state vào Redis.
- Gộp event theo slot.
- Batch update MySQL.
- Gửi event realtime cho frontend.

Worker có thể scale ngang bằng cách chia topic theo `lot_id` hoặc dùng consumer group nếu chuyển MQTT message sang queue/stream phía backend.

**Redis Latest State**

Redis lưu trạng thái mới nhất của slot để frontend/dashboard đọc nhanh và để worker so sánh trước khi ghi MySQL.

Key ví dụ:

```text
slot_state:{lot_id}:{device_mac}:{port}
```

Value ví dụ:

```json
{
  "is_occupied": true,
  "status": "OCCUPIED",
  "changed_at": "2026-05-12T10:29:58+07:00",
  "sequence": 1024
}
```

**MySQL**

MySQL vẫn lưu dữ liệu bền vững. Worker ghi vào MySQL theo batch, không ghi từng MQTT message một cách máy móc.

**ParkingHub / WebTransport**

Worker hoặc service realtime lấy thay đổi đã xử lý và push sang frontend. Frontend nhận trạng thái slot sau khi backend đã lọc trùng và cập nhật cache.

**HTTP IoT API cho Gate/RFID**

Gate/RFID vẫn có thể dùng HTTP:

```text
ESP32 Gate + Camera/RFID
  -> HTTP IoT API
  -> Redis PlateCache
  -> MySQL session / wallet
  -> Response open_barrier / reject
```

Lý do giữ HTTP cho gate là barrier cần câu trả lời ngay. Nếu dùng MQTT cho gate, phải thiết kế request-response qua topic command, phức tạp hơn và cần timeout/correlation id.

### Sensor batch gửi ra sao qua MQTT

Topic:

```text
parking/1/sensor/SENSOR_A_001/batch
```

Payload:

```json
{
  "device_mac": "SENSOR_A_001",
  "lot_id": 1,
  "sequence": 2048,
  "sent_at": "2026-05-12T10:30:00+07:00",
  "events": [
    {
      "port": 1,
      "is_occupied": true,
      "changed_at": "2026-05-12T10:29:57+07:00"
    },
    {
      "port": 2,
      "is_occupied": true,
      "changed_at": "2026-05-12T10:29:58+07:00"
    },
    {
      "port": 5,
      "is_occupied": false,
      "changed_at": "2026-05-12T10:29:59+07:00"
    }
  ]
}
```

MQTT QoS đề xuất:

- QoS 0 nếu chấp nhận mất một vài status update vì batch sau hoặc heartbeat sẽ sửa lại trạng thái.
- QoS 1 nếu muốn đảm bảo broker nhận ít nhất một lần.

Nếu dùng QoS 1, backend phải idempotent vì message có thể bị gửi lại. `device_mac + sequence` hoặc `device_mac + port + changed_at` có thể dùng để chống xử lý trùng.

### Khi nào sensor gửi lên

Giống Option 1, sensor không gửi theo từng lần đo raw distance. Sensor gửi batch khi:

- Có ít nhất một slot đổi trạng thái sau debounce.
- Batch đạt kích thước tối đa, ví dụ 8-32 events.
- Hết thời gian batch window, ví dụ 1-5 giây.
- Có heartbeat định kỳ, ví dụ 30-60 giây, để backend biết thiết bị còn hoạt động.

Ví dụ với một board quản lý 8 slot:

- Nếu trong 5 giây có 3 slot đổi trạng thái, ESP32 publish 1 MQTT message chứa 3 events.
- Nếu 8 slot đổi gần như cùng lúc, ESP32 publish 1 message chứa 8 events.
- Nếu không có slot nào đổi, ESP32 không publish status batch, chỉ gửi heartbeat thưa.

### Dữ liệu đi như thế nào

Luồng sensor:

```text
ESP32 Sensors
  -> MQTT Publish Batch
  -> MQTT Broker
  -> Sensor Worker Subscribe
  -> Redis Latest State
  -> MySQL batch update
  -> ParkingHub / WebTransport
  -> Frontend
```

Luồng gate:

```text
ESP32 Gate + Camera/RFID
  -> HTTP IoT API
  -> Redis PlateCache
  -> MySQL parking_sessions / wallet
  -> HTTP response open_barrier / reject
  -> Servo / LCD / LED
```

### Vì sao xử lý được concurrency lớn

Option 2 xử lý concurrency tốt hơn Option 1 ở tầng thiết bị vì MQTT sinh ra cho mô hình nhiều thiết bị kết nối lâu dài. Thay vì 10.000 thiết bị cùng tạo 10.000 HTTP request ngắn hạn, thiết bị giữ kết nối MQTT và publish message nhẹ hơn.

Khi 10.000 sensor gửi batch gần cùng lúc:

- MQTT broker nhận message từ nhiều thiết bị và làm lớp đệm đầu tiên.
- Backend worker subscribe theo tốc độ có thể kiểm soát.
- Worker gộp event theo slot trước khi ghi MySQL.
- Redis giữ latest state để đọc/ghi nhanh.
- MySQL chỉ nhận batch update đã lọc trùng.
- Có thể scale worker theo `lot_id`, theo topic partition hoặc theo nhiều consumer.

Option 2 tách rõ ingestion layer khỏi processing layer. Broker chịu tải kết nối IoT, worker chịu xử lý nghiệp vụ sensor, MySQL chỉ chịu dữ liệu đã được làm sạch và gom lô.

### Ưu điểm và hạn chế

Ưu điểm:

- Phù hợp IoT thực tế hơn HTTP.
- Giảm overhead kết nối khi số lượng thiết bị lớn.
- Dễ mở rộng bằng broker cluster.
- Hỗ trợ QoS, retained message, last will, device online/offline.
- Sensor batch giúp giảm số message và giảm áp lực worker.

Hạn chế:

- Phải triển khai thêm MQTT broker.
- Firmware phức tạp hơn HTTP đơn giản.
- Cần thiết kế authentication topic, ACL, reconnect, QoS và duplicate handling.

## So sánh nhanh

| Tiêu chí | Option 1: HTTP Batch + Redis Worker | Option 2: MQTT Batch + Worker |
| --- | --- | --- |
| Giao thức sensor | HTTP batch | MQTT batch |
| Mức thay đổi firmware | Trung bình | Cao hơn |
| Mức phù hợp IoT lớn | Tốt | Rất tốt |
| Khả năng chịu burst | Redis queue hấp thụ | MQTT broker + Redis hấp thụ |
| Ghi MySQL | Batch qua worker | Batch qua worker |
| Gate/RFID | HTTP đồng bộ | HTTP đồng bộ |
| Scale backend | Load balancer + worker | Broker + worker theo topic |
| Độ phức tạp vận hành | Thấp hơn | Cao hơn |

## Khuyến nghị

Nếu mục tiêu là nâng cấp từ code hiện tại với rủi ro thấp, chọn Option 1. Đây là bước chuyển hợp lý vì vẫn giữ HTTP, chỉ thay đổi từ single event sang batch và thêm Redis/worker để chống nghẽn MySQL.

Nếu mục tiêu là thuyết phục theo hướng hệ thống thực tế, có khả năng mở rộng tốt cho nhiều bãi xe và nhiều thiết bị, chọn Option 2. Kiến trúc này đúng bản chất IoT hơn: sensor gửi telemetry batch qua MQTT, broker chịu tải kết nối, backend worker xử lý theo batch, Redis giữ trạng thái mới nhất, MySQL lưu dữ liệu bền vững.

Trong cả 2 option, điểm quan trọng nhất là không để sensor request ghi MySQL trực tiếp theo kiểu blocking. Sensor chỉ đẩy dữ liệu vào tầng ingest nhanh, còn việc xử lý, lọc trùng, ghi database và push realtime được thực hiện bất đồng bộ bởi worker.
