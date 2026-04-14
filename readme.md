<div align="center">

# 🛒 Go E-Commerce Platform
### 基于 Go 的云原生微服务电商平台

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.25.5-00ADD8?style=for-the-badge&logo=go&logoColor=white" />
  <img src="https://img.shields.io/badge/go--zero-1.9.4-blue?style=for-the-badge" />
  <img src="https://img.shields.io/badge/gRPC-1.78.0-244c5a?style=for-the-badge&logo=grpc&logoColor=white" />
  <img src="https://img.shields.io/badge/License-MIT-green?style=for-the-badge" />
</p>

<p align="center">
  <img src="https://img.shields.io/badge/MySQL-8.0-4479A1?style=flat-square&logo=mysql&logoColor=white" />
  <img src="https://img.shields.io/badge/Redis-7-DC382D?style=flat-square&logo=redis&logoColor=white" />
  <img src="https://img.shields.io/badge/Kafka-7.5-231F20?style=flat-square&logo=apachekafka&logoColor=white" />
  <img src="https://img.shields.io/badge/Elasticsearch-8.15-005571?style=flat-square&logo=elasticsearch&logoColor=white" />
  <img src="https://img.shields.io/badge/MongoDB-7-47A248?style=flat-square&logo=mongodb&logoColor=white" />
  <img src="https://img.shields.io/badge/etcd-3.5.15-419EDA?style=flat-square" />
  <img src="https://img.shields.io/badge/Prometheus-2.54-E6522C?style=flat-square&logo=prometheus&logoColor=white" />
  <img src="https://img.shields.io/badge/OpenTelemetry-Jaeger-7B5EA7?style=flat-square&logo=opentelemetry&logoColor=white" />
  <img src="https://img.shields.io/badge/Docker-Compose-2496ED?style=flat-square&logo=docker&logoColor=white" />
</p>

<p align="center">
  <strong>16 Microservices · Event-Driven · High-Performance Flash Sales · Full-Text Search · Full Observability</strong><br/>
  <strong>16 个微服务 &nbsp;·&nbsp; 事件驱动 &nbsp;·&nbsp; 高性能秒杀 &nbsp;·&nbsp; 全文搜索 &nbsp;·&nbsp; 可观测性全覆盖</strong>
</p>

---

[English](#-overview) &nbsp;·&nbsp; [中文](#-项目简介) &nbsp;·&nbsp; [Quick Start · 快速开始](#-quick-start--快速开始) &nbsp;·&nbsp; [Architecture · 架构图](#-architecture--系统架构) &nbsp;·&nbsp; [API Docs · 文档](#-api-documentation--api-文档)

</div>

---

## 📖 Overview

A **production-grade cloud-native e-commerce platform** built with Go, implementing a complete microservices architecture. The system handles everything from user authentication to flash-sale order processing, with a focus on high availability, horizontal scalability, and end-to-end observability.

## 📖 项目简介

一个使用 Go 语言构建的**生产级云原生电商平台**，实现了完整的微服务架构。系统覆盖从用户认证到秒杀抢购的全业务链路，专注于高可用、横向扩展与可观测性。

---

## ✨ Features · 核心特性

| Feature | Description | 功能 | 说明 |
|---------|-------------|------|------|
| 🏗️ **Microservices** | 16 independent domain services | **微服务架构** | 16 个独立领域服务 |
| ⚡ **Flash Sales** | Redis + Kafka high-throughput seckill engine | **高性能秒杀** | Redis + Kafka 秒杀引擎 |
| 🔍 **Full-Text Search** | Elasticsearch-powered product search | **全文搜索** | Elasticsearch 商品搜索 |
| 📦 **Event-Driven** | Outbox pattern + Kafka async event streams | **事件驱动** | Outbox 模式 + Kafka 消息流 |
| 🛡️ **JWT Auth** | Token-based authentication & interceptors | **JWT 鉴权** | 基于 Token 的身份认证 |
| 📊 **Observability** | Prometheus metrics + Jaeger + Zipkin tracing | **全链路监控** | 指标采集 + 分布式追踪 |
| 🔄 **Service Discovery** | etcd-based dynamic service routing | **服务发现** | 基于 etcd 的动态路由 |
| 🗄️ **Multi-Database** | MySQL + Redis + MongoDB + Elasticsearch | **多数据库** | 多存储引擎按场景选用 |
| 📡 **gRPC + Protobuf** | Strongly-typed service contracts | **gRPC 通信** | Protobuf 强类型服务契约 |
| 🕷️ **Data Crawler** | Selenium-based product data seeder | **数据爬虫** | 基于 Selenium 的商品数据采集 |

---

## 🏛️ Architecture · 系统架构

```
┌──────────────────────────────────────────────────────────────────────────┐
│                          Clients (Browser / App)                         │
└─────────────────────────────────┬────────────────────────────────────────┘
                                  │  HTTP / REST
┌─────────────────────────────────▼────────────────────────────────────────┐
│                         API Gateway  :8080                               │
│           HTTP → gRPC Proxy  ·  JWT Auth  ·  CORS  ·  File Upload       │
│                         Swagger UI  :8095                                │
└──┬──────┬──────┬──────┬──────┬──────┬──────┬──────┬──────┬──────────────┘
   │      │      │      │      │      │      │      │      │   gRPC / Protobuf
   ▼      ▼      ▼      ▼      ▼      ▼      ▼      ▼      ▼
 User  Product Order Payment  Cart Seckill Search Review  ...
:8000  :8081  :8082  :8083  :8085  :8090  :8010  :8007
   │      │      │      │      │      │      │      │
   └──────┴──────┴──┬───┴──────┴──────┴──────┴──────┘
                    │  Service-to-Service gRPC (etcd discovery)
┌───────────────────▼──────────────────────────────────────────────────────┐
│                          Infrastructure Layer                            │
│                                                                          │
│  ┌──────────┐  ┌──────────┐  ┌───────────────┐  ┌────────────────────┐ │
│  │ MySQL 8  │  │ Redis 7  │  │  Apache Kafka  │  │  Elasticsearch 8   │ │
│  │ Primary  │  │ Cache ·  │  │  Event Stream  │  │   Search Index     │ │
│  │  RDBMS   │  │ Seckill  │  │  Async Jobs    │  │   ES Indexing      │ │
│  └──────────┘  └──────────┘  └───────────────┘  └────────────────────┘ │
│                                                                          │
│  ┌──────────┐  ┌──────────┐  ┌───────────────┐  ┌────────────────────┐ │
│  │ MongoDB  │  │   etcd   │  │  Prometheus   │  │  Jaeger / Zipkin   │ │
│  │  Logs &  │  │ Service  │  │   Metrics     │  │  Distributed       │ │
│  │Analytics │  │Discovery │  │   Scraping    │  │  Tracing           │ │
│  └──────────┘  └──────────┘  └───────────────┘  └────────────────────┘ │
└──────────────────────────────────────────────────────────────────────────┘
```

### Event-Driven Flow · 事件流转

```
Order Created  ──→  outbox_event  ──→  Kafka Topic      ──→  Inventory / Payment
Flash Sale     ──→  Redis DECR    ──→  Kafka Queue       ──→  Async Order Consumer
Product Update ──→  Kafka Topic   ──→  ES Indexer        ──→  Search Index Updated
```

---

## 🗂️ Project Structure · 项目结构

```
go-ecom/
├── api/                         # Protobuf service definitions (16 services)
│   ├── user/v1/
│   ├── product/v1/
│   ├── order/v1/
│   ├── seckill/v1/
│   └── ...
├── cmd/                         # Service entry points (16 services + crawler)
│   ├── api-gateway/             # HTTP → gRPC gateway (:8080)
│   ├── user-service/            # User & auth (:8000)
│   ├── product-service/         # Product catalog (:8081)
│   ├── order-service/           # Orders (:8082)
│   ├── payment-service/         # Payments (:8083)
│   ├── inventory-service/       # Stock management (:8084)
│   ├── cart-service/            # Shopping cart (:8085)
│   ├── promotion-service/       # Coupons & discounts (:8006)
│   ├── review-service/          # Product reviews (:8007)
│   ├── logistics-service/       # Shipment tracking (:8008)
│   ├── message-service/         # Notifications (:8009)
│   ├── search-service/          # Elasticsearch search (:8010)
│   ├── recommend-service/       # Recommendations (:8011)
│   ├── file-service/            # File upload/serve (:8012)
│   ├── job-service/             # Scheduled tasks (:8013)
│   ├── seckill-service/         # Flash sales (:8090)
│   ├── order-service-consumer/  # Kafka async consumer (seckill)
│   └── mi-crawler/              # Python + Selenium data crawler
├── internal/
│   ├── handler/                 # HTTP request handlers
│   ├── middleware/              # JWT auth, CORS, Prometheus
│   ├── pkg/                     # Shared packages
│   │   ├── cache/               # Redis client wrappers
│   │   ├── database/            # MySQL / MongoDB connectors
│   │   ├── mq/                  # Kafka producer & consumer
│   │   └── logger/              # Structured logging (zap)
│   └── service/                 # Business logic (per domain)
├── configs/
│   ├── dev/                     # Development configs
│   ├── test/                    # Test configs
│   └── prod/                    # Production configs
├── database/
│   └── schema.sql               # MySQL schema (30+ tables)
├── docs/swagger/                # OpenAPI / Swagger specs
├── scripts/                     # Setup & deployment scripts
├── docker-compose-infra.yml     # All infrastructure containers
├── prometheus.yml               # Prometheus scrape config
└── Makefile                     # 50+ build & management targets
```

---

## 🧩 Microservices · 微服务一览

| Service | Port | Domain | Key Highlights |
|---------|------|--------|----------------|
| **API Gateway** | 8080 | Routing | HTTP→gRPC, CORS, File upload (100MB max) |
| **User** | 8000 | Auth | JWT, bcrypt, multi-login (username / phone / email) |
| **Product** | 8081 | Catalog | SPU/SKU model, category tree, Redis cache |
| **Order** | 8082 | Commerce | Order state machine, gRPC fan-out (user/product/inventory/logistics/promotion), Saga cancel |
| **Payment** | 8083 | Finance | idgen payment/refund numbers, Kafka callback → order state update |
| **Inventory** | 8084 | Stock | Lua atomic stock deduction, lock/unlock/rollback, Kafka consumer sync to MySQL |
| **Cart** | 8085 | UX | Upsert add-item, SKU status + stock validation before add, price-enriched GetCart |
| **Promotion** | 8006 | Marketing | Atomic coupon receive (SQL WHERE guard, no TOCTOU), discount calculation |
| **Review** | 8007 | Social | Order status gate (must be completed=4, ownership check), star ratings, replies |
| **Logistics** | 8008 | Fulfillment | idgen logistics numbers, shipment creation triggered by ShipOrder |
| **Message** | 8009 | Notify | Kafka consumer: order.created / cancelled / payment.success / refunded → notifications |
| **Search** | 8010 | Discovery | Elasticsearch indexing, real-time sync, faceted search |
| **Recommend** | 8011 | AI/ML | Redis ZSet personalized / hot / similar / realtime recommendations |
| **File** | 8012 | Storage | Image upload/serve, batch support, CORS-enabled |
| **Job** | 8013 | Async | Expired order cancel (UnlockStock first), coupon expiry, low-stock alerts |
| **Seckill** | 8090 | Flash Sale | Redis stock deduction, Kafka queue, rate limiting |
| **Order Consumer** | — | Async | Kafka consumer for seckill order async processing |

---

## 🗄️ Tech Stack · 技术栈

<table>
<tr>
<td valign="top" width="50%">

### Backend · 后端核心
| Category | Technology |
|----------|-----------|
| Language | Go 1.25.5 |
| Framework | go-zero v1.9.4 |
| RPC | gRPC v1.78.0 + Protobuf |
| ORM | GORM v1.31 (MySQL driver) |
| Auth | JWT (golang-jwt/v5) + bcrypt |
| UUID | google/uuid v1.6 |
| Rate Limit | golang.org/x/time |

### Data Storage · 数据存储
| Category | Technology |
|----------|-----------|
| Primary DB | MySQL 8.0 |
| Cache | Redis 7 |
| Search Engine | Elasticsearch 8.15 |
| Document Store | MongoDB 7 |

</td>
<td valign="top" width="50%">

### Infrastructure · 基础设施
| Category | Technology |
|----------|-----------|
| Message Queue | Apache Kafka 7.5 |
| Service Discovery | etcd v3.5.15 |
| Containerization | Docker Compose |
| Coordination | Zookeeper 7.5 |

### Observability · 可观测性
| Category | Technology |
|----------|-----------|
| Metrics | Prometheus v2.54 |
| Tracing | OpenTelemetry + Jaeger |
| Alt Tracing | Zipkin |
| Profiling | Grafana Pyroscope |

### Dev Tools · 开发工具
| Category | Technology |
|----------|-----------|
| API Docs | Swagger / OpenAPI |
| gRPC Debug | grpcurl + gRPC Reflection |
| Data Seeding | Python + Selenium |
| Build | Makefile (50+ targets) |

</td>
</tr>
</table>

---

## 🚀 Quick Start · 快速开始

### Prerequisites · 前置要求

- **Go** 1.21+
- **Docker** & **Docker Compose**
- **MySQL** 8.0 (local or via Docker)
- **protoc** (for protobuf code generation, optional)

---

### Step 1 · Clone & Setup

```bash
git clone https://github.com/your-username/go-ecom.git
cd go-ecom

# Download Go dependencies · 下载依赖
make deps
```

---

### Step 2 · Start Infrastructure · 启动基础设施

```bash
# Start all infrastructure containers
# 启动所有基础设施容器
make start-infra

# Verify all containers are healthy · 验证容器状态
docker compose -f docker-compose-infra.yml ps
```

| Service | URL | Purpose |
|---------|-----|---------|
| Redis | `localhost:6379` | Cache & seckill stock |
| Kafka | `localhost:9092` | Event streaming |
| Kafka UI | `http://localhost:18090` | Kafka web management |
| Elasticsearch | `http://localhost:9200` | Search engine |
| etcd | `localhost:2379` | Service discovery |
| etcdkeeper | `http://localhost:8089` | etcd web UI |
| MongoDB | `localhost:27017` | Document store |
| Prometheus | `http://localhost:9090` | Metrics dashboard |

---

### Step 3 · Initialize Database · 初始化数据库

```bash
# Apply the schema · 导入数据库 Schema
mysql -u root -p ecommerce < database/schema.sql
```

---

### Step 4 · Configure Services · 配置服务

Edit `configs/dev/*.yaml` for your local environment.

编辑 `configs/dev/` 下各服务配置文件：

```yaml
# configs/dev/user-config.yaml (example)
Database:
  Host: localhost
  Port: 3306
  User: root
  Password: "your_password"
  Database: ecommerce

BizRedis:
  Host: localhost
  Port: 6379

JWT:
  Secret: your-secret-key
  Expire: 7200
```

---

### Step 5 · Start Backend · 启动后端

```bash
# Start API gateway + all backend services
# 启动 API 网关 + 全部后端服务
make start-backend
```

### Step 6 · Start Frontend · 启动前端

```bash
# Start both frontends in background
# 后台启动两个前端
make start-frontend
```

启动完成后可访问：

| App | URL | 说明 |
|---|---|---|
| API Gateway | `http://localhost:8080` | 前后端联调统一入口 |
| Swagger UI | `http://localhost:8095` | API 文档 |
| Frontend User | `http://localhost:5173` | 商城前台 |
| Frontend Admin | `http://localhost:5174` | 管理后台 |

说明：

- 两个前端都已配置代理到 `http://localhost:8080`
- `make start-frontend` 会自动检查并安装前端依赖
- 前端日志输出到 `logs/frontend-user.log` 和 `logs/frontend-admin.log`
- 文件上传走主 HTTP 服务的 `/api/v1/files/upload` 和 `/api/v1/files/batch-upload`
- 这些上传接口不经过 gRPC Gateway 映射，而是由 `cmd/api-gateway` 中注册的 HTTP handler 直接处理

---

### Step 7 · Seed Data (Optional) · 填充测试数据（可选）

```bash
cd cmd/mi-crawler

# Install Python dependencies · 安装 Python 依赖
pip install -r requirements.txt

# Run the 3-phase crawler pipeline · 运行三阶段爬虫流水线
python main.py --phase categories   # Crawl categories · 抓取分类
python main.py --phase products     # Crawl products  · 抓取商品
python main.py --phase skus         # Crawl SKUs      · 抓取 SKU
```

---

## ⚡ Flash Sale (Seckill) · 秒杀功能

### One-command full setup · 一键启动

```bash
make seckill-full
```

### Manual steps · 手动步骤

```bash
# Set stock for SKU in Redis · 设置 Redis 秒杀库存
make redis-set-stock SKU_ID=1 STOCK=100

# Query stock · 查询库存
make redis-get-stock SKU_ID=1

# List all stocks · 查看所有库存
make redis-list-stocks

# Stop seckill services · 停止秒杀服务
make seckill-stop
```

### How it works · 工作原理

```
User Request
     │
     ▼
API Gateway  ──→  Seckill Service (:8090)
                        │
             ┌──────────▼───────────┐
             │  Redis DECR           │   ← Atomic O(1) stock deduction
             │  seckill:stock:{id}   │     原子操作，O(1) 扣减库存
             └──────────┬───────────┘
                        │ success
             ┌──────────▼───────────┐
             │  Kafka Producer       │   ← Async queue absorbs traffic spikes
             │  topic: seckill-order │     异步队列吸收流量峰值
             └──────────┬───────────┘
                        │
             ┌──────────▼───────────┐
             │  Order Consumer       │   ← Creates order + locks inventory
             │  (Async Processing)   │     异步创建订单并锁定库存
             └───────────────────────┘
```

---

## 📡 API Documentation · API 文档

- **Swagger UI**: `http://localhost:8095`
- **Base URL**: `http://localhost:8080/api/v1`

<details>
<summary><strong>👤 User Service · 用户服务</strong></summary>

```
POST   /api/v1/user/register        # Register            · 注册
POST   /api/v1/user/login           # Login               · 登录
GET    /api/v1/user/info            # Get user profile     · 获取用户信息
PUT    /api/v1/user/info            # Update profile       · 更新用户信息
GET    /api/v1/user/address         # List addresses       · 地址列表
POST   /api/v1/user/address         # Add address          · 新增地址
PUT    /api/v1/user/address/:id     # Update address       · 更新地址
```
</details>

<details>
<summary><strong>📦 Product Service · 商品服务</strong></summary>

```
GET    /api/v1/products             # List products        · 商品列表
GET    /api/v1/products/:id         # Product detail       · 商品详情
GET    /api/v1/products/category/*  # Filter by category   · 按分类查询
```
</details>

<details>
<summary><strong>📋 Order Service · 订单服务</strong></summary>

```
POST   /api/v1/orders               # Create order         · 创建订单
GET    /api/v1/orders               # List user orders     · 我的订单列表
GET    /api/v1/orders/:id           # Order detail         · 订单详情
DELETE /api/v1/orders/:id           # Cancel order         · 取消订单
```
</details>

<details>
<summary><strong>💳 Payment Service · 支付服务</strong></summary>

```
POST   /api/v1/payment              # Initiate payment     · 发起支付
GET    /api/v1/payment/:id          # Payment status       · 支付状态查询
```
</details>

<details>
<summary><strong>🛒 Cart Service · 购物车</strong></summary>

```
GET    /api/v1/cart                 # Get cart             · 查看购物车
POST   /api/v1/cart/items           # Add item             · 加入购物车
PUT    /api/v1/cart/items/:id       # Update quantity      · 修改数量
DELETE /api/v1/cart/items/:id       # Remove item          · 删除商品
```
</details>

<details>
<summary><strong>⚡ Seckill Service · 秒杀服务</strong></summary>

```
GET    /api/v1/seckill/activities   # Flash sale list      · 秒杀活动列表
POST   /api/v1/seckill              # Flash purchase       · 秒杀下单
```
</details>

<details>
<summary><strong>🔍 Search Service · 搜索服务</strong></summary>

```
GET    /api/v1/search?q=keyword     # Search products      · 商品搜索
```
</details>

<details>
<summary><strong>📁 File Service · 文件服务</strong></summary>

```
POST   /api/v1/files/upload         # Upload file          · 上传文件（最大 100MB）
GET    /uploads/*                   # Serve uploaded files · 访问上传文件
GET    /images/*                    # Serve crawler images · 访问爬虫图片
```
</details>

---

## 🔨 Build & Development · 构建与开发

```bash
# Build all services · 构建所有服务
make build

# Build a single service · 构建单个服务
make build-service SERVICE=user-service

# Run tests · 运行测试
make test

# Lint the codebase · 代码静态检查
make lint

# Generate protobuf Go code · 生成 Protobuf Go 代码
make proto

# Generate proto descriptors for the gateway · 生成 Gateway 用 Descriptor 文件
make proto-descriptor

# Generate Swagger / OpenAPI documentation · 生成 Swagger 文档
make swagger

# Clean build artifacts · 清理构建产物
make clean

# Show all available targets · 查看所有可用命令
make help
```

---

## 🗃️ Database Schema · 数据库结构

The MySQL schema (`database/schema.sql`) contains **30+ tables** organized by domain.

数据库 Schema 包含 **30+ 张表**，按领域划分：

| Domain · 领域 | Tables · 表名 |
|---------------|--------------|
| **User · 用户** | `user`, `credential`, `address` |
| **Product · 商品** | `product`, `sku`, `category`, `brand`, `attr` |
| **Inventory · 库存** | `inventory`, `inventory_log` |
| **Order · 订单** | `orders`, `order_item`, `order_log` |
| **Payment · 支付** | `payment`, `payment_log` |
| **Cart · 购物车** | `cart` |
| **Promotion · 促销** | `coupon`, `user_coupon`, `promotion`, `points` |
| **Review · 评价** | `review`, `review_reply` |
| **Logistics · 物流** | `logistics` |
| **Message · 消息** | `message` |
| **System · 系统** | `outbox_event`, `banner`, `system_config` |

### Outbox Pattern · 事务性消息表

```sql
-- Guarantees reliable event delivery without distributed transactions
-- 无需分布式事务，保证可靠的跨服务消息投递
CREATE TABLE outbox_event (
  id         BIGINT PRIMARY KEY AUTO_INCREMENT,
  topic      VARCHAR(128) NOT NULL,    -- Kafka topic name
  payload    JSON NOT NULL,            -- Event payload
  status     TINYINT DEFAULT 0,        -- 0: pending  1: sent  2: failed
  created_at DATETIME DEFAULT NOW()
);
```

---

## 📊 Observability · 可观测性

### Metrics · 指标监控

Each microservice exposes `/metrics` on a dedicated port. Prometheus scrapes all of them.

每个微服务在独立端口暴露 `/metrics`，Prometheus 统一采集：

| Service | Metrics Port |
|---------|-------------|
| User | :9091 |
| Product | :9092 |
| Order | :9093 |
| ... | ... |
| **Prometheus UI** | **:9090** |

### Distributed Tracing · 分布式链路追踪

OpenTelemetry is integrated across all services, exporting traces to Jaeger.

全服务集成 OpenTelemetry，链路数据导出到 Jaeger：

```yaml
# Enabled via service config · 通过服务配置开启
Middlewares:
  Trace: true       # OpenTelemetry tracing
  Recover: true     # Panic recovery
  Stat: true        # Request statistics
  Prometheus: true  # Metrics exposition
  Breaker: true     # Circuit breaker
```

---

## 🔧 Shared Infrastructure · 公共基础层

All microservices are built on top of a unified shared package layer at `internal/pkg/`. This eliminates copy-paste across services and ensures consistent behaviour.

所有微服务共享同一套基础包，消除跨服务重复代码，确保行为一致：

### `internal/pkg/errors` — Unified Error Codes

Central gRPC status code mapping. Each service calls `errors.ConvertToGRPCError(err)` — no per-service `convertError` helpers.

统一错误码映射。各服务直接调用 `errors.ConvertToGRPCError(err)`，不再各自维护转换函数。

### `internal/pkg/cache` — Redis Helpers

| Function | Purpose |
|----------|---------|
| `AtomicDeductStock(ctx, key, qty)` | Lua-based atomic stock deduction (no TOCTOU) |
| `AtomicRollbackStock(ctx, key, qty)` | Rollback a failed deduction |
| `IsNil(err)` | Safe `redis.Nil` check (avoids direct import coupling) |
| `SetNX(ctx, key, val, ttl)` | Distributed lock primitive |

### `internal/pkg/idgen` — Business Number Generator

Redis `INCR` + date prefix, falls back to nanosecond timestamp when Redis is unavailable.

| Method | Format Example |
|--------|---------------|
| `OrderNo(ctx)` | `ORD20260414000001` |
| `PaymentNo(ctx)` | `PAY20260414000001` |
| `RefundNo(ctx)` | `REF20260414000001` |
| `LogisticsNo(ctx)` | `LOG20260414000001` |

### `internal/pkg/middleware` — gRPC Auth Interceptor

`AuthInterceptor` / `RequireAuthInterceptor` — shared JWT validation interceptor reused by all services. Removed per-service `interceptor/` directories in user and cart.

统一 JWT 校验拦截器，各服务复用，删除原 user/cart 各自的 interceptor 目录。

### `internal/pkg/client` — Typed gRPC Clients

Pre-wired, timeout-aware gRPC clients for each downstream service. All use a unified `RpcConf` struct with 5-second default timeout and retry.

| Client | Key Methods |
|--------|-------------|
| `UserClient` | `GetUser`, `GetUserAddress` |
| `ProductClient` | `GetSKU`, `GetProduct` |
| `InventoryClient` | `LockStock`, `UnlockStock`, `DeductStock` |
| `OrderClient` | `GetOrder`, `PayOrder`, `CancelOrder`, `RefundOrder`, `ShipOrder` |
| `LogisticsClient` | `CreateLogistics` |
| `PromotionClient` | `CalculateDiscount`, `UseCoupon` |

---

## 🧱 Design Patterns · 设计模式

| Pattern | Where Used | 模式 | 应用场景 |
|---------|-----------|------|---------|
| **Repository** | All services | **仓储模式** | 所有服务数据访问层解耦 |
| **Outbox** | Order / Payment | **事务性发件箱** | 订单、支付可靠消息投递 |
| **Cache-Aside** | Product / Cart | **旁路缓存** | 商品与购物车读写缓存 |
| **CQRS** | Search | **读写分离** | ES 读、MySQL 写分离 |
| **Circuit Breaker** | All RPC calls | **熔断器** | 所有 gRPC 调用容错保护 |
| **API Gateway** | Entry point | **统一网关** | 单一入口路由与鉴权 |
| **Saga** | Order flow | **Saga 模式** | 跨服务订单创建流程 |
| **Service Context** | All services | **依赖注入容器** | 统一管理 DB/Redis/Kafka 连接 |
| **Atomic SQL Guard** | Promotion coupon | **原子 SQL 防超发** | `UPDATE ... WHERE used_count < total_count`，无需应用层加锁 |
| **Lua Atomic Script** | Inventory stock | **Lua 原子扣减** | Redis Lua 脚本保证扣减+校验原子性，消除 TOCTOU |
| **Event Consumer** | Message / Inventory | **Kafka 消费者** | 后台 goroutine 消费订单/支付事件，异步发通知或同步库存 |
| **Fail-Fast Init** | All services | **启动强校验** | DB/Redis 连接失败直接 `log.Fatal`，不静默放行 |

---

## 📁 Configuration · 配置说明

Multi-environment YAML configuration, loaded at startup with `-f <path>`.

多环境 YAML 配置，启动时通过 `-f <配置路径>` 指定：

```
configs/
├── dev/        # 开发环境 / Development
│   ├── user-config.yaml
│   ├── product-config.yaml
│   ├── order-config.yaml
│   └── ...
├── test/       # 测试环境 / Test
└── prod/       # 生产环境 / Production
```

All services support three run modes: `dev` (gRPC Reflection enabled) · `test` · `prod`.

所有服务支持三种运行模式：`dev`（开启 gRPC Reflection）· `test` · `prod`。

---

## 🤝 Contributing · 贡献指南

1. Fork the repository
2. Create a feature branch: `git checkout -b feat/your-feature`
3. Commit your changes following [Conventional Commits](https://www.conventionalcommits.org/)
4. Push to your fork and open a Pull Request

欢迎提交 Issue 和 PR！请遵循 [Conventional Commits](https://www.conventionalcommits.org/) 规范。

---

## 📄 License · 许可证

This project is licensed under the **MIT License**.

本项目基于 **MIT 许可证** 开源，详见 [LICENSE](LICENSE)。

---

<div align="center">
<sub>Built with ❤️ in Go &nbsp;·&nbsp; gRPC &nbsp;·&nbsp; Kafka &nbsp;·&nbsp; Redis &nbsp;·&nbsp; Elasticsearch</sub>
</div>
