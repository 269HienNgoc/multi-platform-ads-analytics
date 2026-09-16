# Báo cáo API backend

Ngày rà soát: 16/09/2026
Phiên bản API: `v1`
Base URL local: `http://127.0.0.1:8080`

## 1. Tóm tắt hiện trạng

Backend hiện đã cung cấp nền tảng cho danh mục quảng cáo đa nền tảng theo mô hình:

`Platform → Ad Account → Campaign → Ad Group/Ad Set → Ad → Creative`

API đang chạy trên Gin, dữ liệu được lưu bằng GORM/PostgreSQL, cấu hình đọc từ YAML qua Viper và log có cấu trúc bằng Zap. Mã định danh nội bộ là UUID; mã của Meta, TikTok hoặc Google chỉ được lưu trong `external_id`.

Hiện có **16 operation**: 2 operation kiểm tra sức khỏe, 7 operation catalog, 4 operation workflow và 3 operation connector Meta. Connector Meta hiện cố ý chạy ở chế độ **read-only**: đồng bộ tài khoản/campaign về PostgreSQL, archive các bản ghi do provider quản lý khi chúng biến mất khỏi một lần đọc đầy đủ thành công, nhưng chưa phát hành campaign hoặc thay đổi ngân sách thật.

## 2. Danh sách endpoint

| Method | Endpoint | Chức năng | Thành công |
| --- | --- | --- | --- |
| `GET` | `/health/live` | Kiểm tra process API đang chạy | `200` |
| `GET` | `/health/ready` | Kiểm tra API và PostgreSQL sẵn sàng | `200` |
| `GET` | `/api/v1/ad-accounts` | Liệt kê tài khoản và số campaign đã đồng bộ | `200` |
| `POST` | `/api/v1/ad-accounts` | Tạo tài khoản quảng cáo | `201` |
| `POST` | `/api/v1/campaigns` | Tạo chiến dịch trong tài khoản | `201` |
| `POST` | `/api/v1/ad-groups` | Tạo nhóm quảng cáo/ad set | `201` |
| `POST` | `/api/v1/ads` | Tạo quảng cáo | `201` |
| `POST` | `/api/v1/creatives` | Tạo nội dung quảng cáo | `201` |
| `GET` | `/api/v1/ad-accounts/{accountID}/hierarchy` | Đọc toàn bộ cây dữ liệu của tài khoản | `200` |
| `GET` | `/api/v1/workflows` | Liệt kê workflow automation | `200` |
| `POST` | `/api/v1/workflows/bulk` | Tạo một workflow cho mỗi tài khoản đã chọn | `201` |
| `POST` | `/api/v1/workflows/{workflowID}/transition` | Chuyển trạng thái workflow hợp lệ | `200` |
| `POST` | `/api/v1/workflows/{workflowID}/metrics` | Ghi metric và tự chuyển Camp mồi đã đủ ngưỡng | `200` |
| `GET` | `/api/v1/connectors/meta` | Xem trạng thái connector Meta | `200` |
| `POST` | `/api/v1/connectors/meta/sync` | Đồng bộ mọi Meta account/campaign nhìn thấy | `200` |
| `GET` | `/api/v1/connectors/meta/sync-runs` | Xem lịch sử đồng bộ | `200` |

## 3. Quy ước chung

### Header

- Request gửi JSON cần có `Content-Type: application/json`.
- Các route `/api/v1/*` yêu cầu `X-API-Key` khi `server.api_key` được cấu hình trong YAML; production bắt buộc cấu hình khóa này.
- `POST /workflows/bulk` nhận `Idempotency-Key`. Nếu client không gửi, backend dùng `X-Request-ID` làm khóa cho lần gọi đó.
- Client có thể gửi `X-Request-ID` gồm 1–64 ký tự chữ, số, `_` hoặc `-`.
- Nếu `X-Request-ID` không hợp lệ hoặc bị thiếu, backend tự sinh ID mới.
- Mọi response đều trả lại `X-Request-ID` để tra log.
- JSON body bị giới hạn ở 1 MiB.

### Giá trị chuẩn hóa

| Trường | Giá trị hợp lệ |
| --- | --- |
| `platform` | `meta`, `tiktok`, `google` |
| `status` | `active`, `paused`, `archived` |
| `currency` | Mã tiền tệ gồm đúng 3 ký tự; backend tự chuyển sang chữ hoa |
| ID cha | UUID nội bộ do backend trả về, không dùng ID của nền tảng |
| `provider_data` | JSON object chứa trường riêng của từng nền tảng; mặc định `{}` |

### Lỗi chuẩn

```json
{
  "error": {
    "code": "invalid_request",
    "message": "invalid catalog data"
  }
}
```

| HTTP | `code` | Khi nào xảy ra |
| --- | --- | --- |
| `400` | `invalid_request` | JSON sai, thiếu trường bắt buộc, UUID sai hoặc dữ liệu không hợp lệ |
| `401` | `unauthorized` | Thiếu hoặc sai `X-API-Key` |
| `404` | `not_found` | Không tìm thấy entity hoặc entity cha |
| `409` | `conflict` | Trùng `external_id` trong cùng phạm vi cha |
| `409` | `invalid_transition` | Yêu cầu chuyển trạng thái workflow không hợp lệ |
| `500` | `internal_error` | Lỗi ngoài dự kiến; chi tiết chỉ ghi vào Zap log |
| `503` | Không dùng error envelope | PostgreSQL chưa sẵn sàng ở `/health/ready` |
| `502` | `provider_sync_failed` | Meta API hoặc bước lưu snapshot thất bại |
| `503` | `connector_unavailable` | Connector Meta chưa được cấu hình |

## 4. Chi tiết endpoint

### `GET /health/live`

Response `200`:

```json
{
  "status": "up",
  "checks": {
    "process": "up"
  }
}
```

### `GET /health/ready`

Response `200` khi PostgreSQL hoạt động:

```json
{
  "status": "up",
  "checks": {
    "database": "up"
  }
}
```

Response `503` khi PostgreSQL không sẵn sàng:

```json
{
  "status": "down",
  "checks": {
    "database": "down"
  }
}
```

### `POST /api/v1/ad-accounts`

Request:

```json
{
  "platform": "meta",
  "external_id": "act_123456789",
  "name": "Meta — Thị trường Việt Nam",
  "currency": "VND",
  "timezone": "Asia/Ho_Chi_Minh",
  "status": "active",
  "provider_data": {
    "business_id": "987654321"
  }
}
```

Các trường bắt buộc: `platform`, `external_id`, `name`, `currency`, `timezone`, `status`.

Response `201`:

```json
{
  "id": "c778ba63-664b-4a57-8ea7-2bd4f8d6d213",
  "platform": "meta",
  "external_id": "act_123456789",
  "name": "Meta — Thị trường Việt Nam",
  "currency": "VND",
  "timezone": "Asia/Ho_Chi_Minh",
  "status": "active",
  "provider_data": {
    "business_id": "987654321"
  },
  "created_at": "2026-09-15T08:00:00Z",
  "updated_at": "2026-09-15T08:00:00Z"
}
```

Phạm vi chống trùng: `platform + external_id`.

### `GET /api/v1/ad-accounts`

Trả `{"data":[...]}` theo thứ tự platform, tên và UUID. Mỗi tài khoản có thêm `campaign_count` và `last_synced_at` để frontend hiển thị dữ liệu đồng bộ thật mà không cần gọi hierarchy cho từng account.

### `POST /api/v1/campaigns`

Request:

```json
{
  "account_id": "c778ba63-664b-4a57-8ea7-2bd4f8d6d213",
  "external_id": "120210001",
  "name": "VN | Purchase | Broad | 09-2026",
  "objective": "sales",
  "status": "active",
  "provider_data": {
    "buying_type": "AUCTION"
  }
}
```

Các trường bắt buộc: `account_id`, `external_id`, `name`, `status`. `objective` và `provider_data` có thể bỏ trống.

Phạm vi chống trùng: `account_id + external_id`.

### `POST /api/v1/ad-groups`

Request:

```json
{
  "campaign_id": "1ed688cc-8eb8-4bcb-b35f-b7a2d1cab063",
  "external_id": "120210002",
  "name": "Broad | VN | 18–45",
  "status": "active",
  "provider_data": {
    "daily_budget": 250000,
    "optimization_goal": "OFFSITE_CONVERSIONS"
  }
}
```

Các trường bắt buộc: `campaign_id`, `external_id`, `name`, `status`.

Phạm vi chống trùng: `campaign_id + external_id`.

### `POST /api/v1/ads`

Request:

```json
{
  "ad_group_id": "cd0e77cf-e09d-42cf-a1f8-70d738e430a8",
  "external_id": "120210003",
  "name": "Video 01 | Hook giá",
  "status": "active",
  "provider_data": {
    "tracking_specs": []
  }
}
```

Các trường bắt buộc: `ad_group_id`, `external_id`, `name`, `status`.

Phạm vi chống trùng: `ad_group_id + external_id`.

### `POST /api/v1/creatives`

Request:

```json
{
  "ad_id": "f8c6bd3f-ab42-4f63-8be7-fd501d4dd58f",
  "external_id": "120210004",
  "name": "UGC 24s — Hook giá",
  "format": "video",
  "asset_url": "https://cdn.example.com/creative-01.mp4",
  "provider_data": {
    "headline": "Ưu đãi hôm nay"
  }
}
```

Các trường bắt buộc: `ad_id`, `external_id`, `name`, `format`. `asset_url` và `provider_data` có thể bỏ trống. Backend chuẩn hóa `format` thành chữ thường.

Phạm vi chống trùng: `ad_id + external_id`.

### `GET /api/v1/ad-accounts/{accountID}/hierarchy`

`accountID` phải là UUID nội bộ. Endpoint trả về toàn bộ campaign, ad group, ad và creative của tài khoản trong một response lồng nhau.

Response `200` rút gọn:

```json
{
  "account": {
    "id": "c778ba63-664b-4a57-8ea7-2bd4f8d6d213",
    "platform": "meta",
    "external_id": "act_123456789",
    "name": "Meta — Thị trường Việt Nam",
    "currency": "VND",
    "timezone": "Asia/Ho_Chi_Minh",
    "status": "active",
    "provider_data": {},
    "created_at": "2026-09-15T08:00:00Z",
    "updated_at": "2026-09-15T08:00:00Z"
  },
  "campaigns": [
    {
      "campaign": {
        "id": "1ed688cc-8eb8-4bcb-b35f-b7a2d1cab063",
        "account_id": "c778ba63-664b-4a57-8ea7-2bd4f8d6d213",
        "external_id": "120210001",
        "name": "VN | Purchase | Broad | 09-2026",
        "objective": "sales",
        "status": "active",
        "provider_data": {},
        "created_at": "2026-09-15T08:02:00Z",
        "updated_at": "2026-09-15T08:02:00Z"
      },
      "ad_groups": [
        {
          "ad_group": {},
          "ads": [
            {
              "ad": {},
              "creatives": []
            }
          ]
        }
      ]
    }
  ]
}
```

### Campaign workflow automation

Tạo workflow hàng loạt bằng `POST /api/v1/workflows/bulk`:

```json
{
  "organization_id": "local-testing",
  "ad_account_ids": ["c778ba63-664b-4a57-8ea7-2bd4f8d6d213"],
  "page_external_id": "1029384756",
  "pixel_external_id": "5647382910",
  "pixel_event": "Purchase",
  "existing_post_id": "1029384756_1234567890",
  "seed_spend_limit_usd": 10
}
```

Mỗi account tạo một record độc lập ở trạng thái `ACCOUNT_CONNECTED`; hệ thống không giả định Page/Pixel/payment đã qua preflight. Một account lỗi không rollback account khác; response dùng `207 Multi-Status` và mảng `failures` nếu chỉ thành công một phần. Retry cùng `organization_id + Idempotency-Key + ad_account_id` trả lại workflow đã tạo thay vì tạo trùng. `GET /api/v1/workflows` trả `{"data": [...]}`. Chuyển trạng thái bằng body `{"state":"ASSETS_SYNCED"}` tại endpoint `transition`.

Endpoint `metrics` nhận `spend_usd`, `registrations`, `deposits`, `cost_per_registration`, `cost_per_deposit` và `captured_at`. Nếu workflow đang `SEED_RUNNING` và `spend_usd` đạt `seed_spend_limit_usd`, backend ghi metric và chuyển nguyên tử sang `MAIN_PENDING`.

### Đồng bộ Meta read-only

Khi `meta.enabled: true` trong YAML, worker chạy ngay lúc khởi động và lặp theo `meta.sync_interval`. Đồng bộ dùng cursor pagination cho tất cả ad account, sau đó lấy campaign của từng account, upsert theo external ID và lưu raw payload để audit. Thiếu quyền Page hoặc Pixel được trả về dưới dạng cảnh báo và không làm mất kết quả Account/Campaign.

Các thao tác tạo campaign, thay ngân sách và đọc insight vẫn fail-closed cho đến khi có preflight, approval và budget guardrail.

## 5. Dữ liệu đã có schema nhưng chưa có API đầy đủ

Migration PostgreSQL đã tạo các bảng sau nhưng router hiện chưa expose endpoint tương ứng:

- `performance_metrics_daily`: impression, reach, click, conversion, spend, revenue và metric riêng của provider theo ngày.
- `raw_provider_payloads`: payload gốc có hash chống trùng để phục vụ audit và AI; đây là dữ liệu nội bộ, không expose ra API.
- `automation_rules`: rule định lượng có ngưỡng chi tiêu/mẫu; domain evaluator đã có nhưng chưa expose CRUD API.

`sync_runs` đã có API đọc riêng cho Meta. Chưa có API tổng hợp lịch sử sync đa nền tảng.

## 6. Khả năng dùng cho frontend hiện tại

| Nhu cầu giao diện | Backend hiện tại | Cách xử lý ở frontend giai đoạn đầu |
| --- | --- | --- |
| Hiển thị trạng thái hệ thống | Đủ | Gọi `/health/ready` qua Next.js Route Handler |
| Thêm tài khoản quảng cáo | Đủ | Gọi `POST /ad-accounts` |
| Xem một cây tài khoản | Đủ khi đã biết UUID | Gọi endpoint `hierarchy` |
| Danh sách tất cả tài khoản | Đủ | Gọi `GET /ad-accounts`; frontend đang dùng dữ liệu thật |
| Dashboard KPI/biểu đồ | Chưa có API đọc metric | Dùng dữ liệu demo có nhãn rõ ràng |
| Cấu hình hàng loạt nhiều tài khoản | Đủ cho workflow foundation | Có partial success và idempotency; chưa publish campaign thật |
| Bật/tắt/sửa/xóa campaign | Chưa có | Chỉ hiển thị trạng thái, không giả lập thao tác ghi |
| Đăng nhập và phân quyền | Một lớp vận hành | Go API dùng API key; Next dashboard dùng HTTP Basic ở production. Chưa có RBAC/tenant identity |

## 7. Khoảng trống cần ưu tiên ở backend

1. Thêm identity, RBAC và tenant/workspace thực sự trước khi mở cho nhiều khách hàng.
2. Thêm phân trang, filter platform/status và tìm kiếm server-side cho `GET /ad-accounts`.
3. Thêm API tổng hợp dashboard theo khoảng ngày và timezone; KPI hiện vẫn là dữ liệu mô phỏng.
4. Hoàn thiện Meta preflight, approval, budget guardrail và publish orchestration trước khi bật tạo campaign thật.
5. Thêm API đọc campaign/ad group/ad theo danh sách thay vì chỉ đọc toàn bộ hierarchy.
6. Bổ sung OpenAPI breaking-change check và integration test PostgreSQL trong CI.

## 8. Bảo mật và vận hành

- API chưa có xác thực/phân quyền nên chỉ nên chạy ở local hoặc private network.
- Middleware đã có request ID, panic recovery, access log Zap và security headers.
- Response lỗi không làm lộ lỗi database hoặc stack trace.
- Cấu hình production có thể giữ PostgreSQL tại `127.0.0.1:5432` nếu backend và database cùng VPS; password được đặt trong file YAML dành riêng cho deployment.
- Không commit file YAML production hoặc thông tin đăng nhập nền tảng quảng cáo; giới hạn quyền đọc file cho tài khoản chạy dịch vụ.

## 9. Nguồn đối chiếu trong mã nguồn

- Router: `backend/internal/adapter/httpapi/router.go`
- Request/response và error mapping: `backend/internal/adapter/httpapi/catalog.go`
- Validation: `backend/internal/application/catalog/service.go`
- Domain constants: `backend/internal/domain/ads/`
- PostgreSQL schema: `backend/migrations/000001_create_ad_catalog.up.sql`
- Automation: `backend/internal/application/automation/` và migration `000002_create_campaign_automation.up.sql`
- Machine-readable contract: `docs/api-reference/openapi.yaml`
