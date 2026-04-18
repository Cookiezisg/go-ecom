# Go E-Commerce Platform

基于 Go、go-zero、gRPC 和多种基础设施组件实现的微服务电商项目。仓库包含后端服务、API Gateway、用户端与管理端前端、Swagger 文档、秒杀链路以及用于灌数的 Python 爬虫。

当前仓库内可确认的组成如下：

- 15 个领域服务的 Proto 定义，位于 `api/*/v1`
- 16 个 Go 可执行入口，位于 `cmd/*`，其中包含 `api-gateway` 和 `order-service-consumer`
- 2 个前端应用：`frontend-user`、`frontend-admin`
- 1 个基础设施编排文件：`docker-compose-infra.yml`
- 1 套开发环境配置：`configs/dev/*.yaml`

## Overview

项目覆盖了典型电商链路：

- 用户注册、登录、地址管理
- 商品、SKU、分类、Banner 管理
- 购物车、订单、支付、库存、优惠券
- 商品评价、物流、消息通知
- 搜索、推荐、文件上传
- 秒杀活动与异步下单消费

核心技术栈：

- Go 1.25.x
- go-zero
- gRPC + Protobuf
- MySQL
- Redis
- Kafka + Zookeeper
- Elasticsearch
- MongoDB
- etcd
- Prometheus
- React + Vite

## Repository Layout

```text
go-ecom/
├── api/                    # Protobuf 定义
├── cmd/                    # 各服务入口、gateway、consumer、crawler
├── configs/dev/            # 当前实际提供的开发配置
├── database/               # schema 与初始化脚本
├── docs/swagger/           # 生成后的 OpenAPI 文档
├── frontend-user/          # 商城前台
├── frontend-admin/         # 管理后台
├── internal/               # 业务实现与共享基础包
├── scripts/                # 启动、检查、生成脚本
├── docker-compose-infra.yml
├── prometheus.yml
└── Makefile
```

## Services

### Backend Service Ports

| Service | Port | Config |
| --- | --- | --- |
| `api-gateway` | `8080` | `configs/dev/gateway.yaml` |
| `user-service` | `8000` | `configs/dev/user-config.yaml` |
| `product-service` | `8081` | `configs/dev/product-config.yaml` |
| `order-service` | `8082` | `configs/dev/order-config.yaml` |
| `payment-service` | `8083` | `configs/dev/payment-config.yaml` |
| `inventory-service` | `8084` | `configs/dev/inventory-config.yaml` |
| `cart-service` | `8085` | `configs/dev/cart-config.yaml` |
| `promotion-service` | `8006` | `configs/dev/promotion-config.yaml` |
| `review-service` | `8007` | `configs/dev/review-config.yaml` |
| `logistics-service` | `8008` | `configs/dev/logistics-config.yaml` |
| `message-service` | `8009` | `configs/dev/message-config.yaml` |
| `search-service` | `8010` | `configs/dev/search-config.yaml` |
| `recommend-service` | `8011` | `configs/dev/recommend-config.yaml` |
| `file-service` | `8012` | `configs/dev/file-config.yaml` |
| `job-service` | `8013` | `configs/dev/job-config.yaml` |
| `seckill-service` | `8090` | `configs/dev/seckill-config.yaml` |
| `order-service-consumer` | - | `configs/dev/order-config.yaml` |

### Frontend Ports

| App | Port |
| --- | --- |
| `frontend-user` | `5173` |
| `frontend-admin` | `5174` |

## Infrastructure

`docker-compose-infra.yml` 当前编排以下依赖：

| Component | Port |
| --- | --- |
| Redis | `6379` |
| etcd | `2379` / `2380` |
| etcdkeeper | `8089` |
| MongoDB | `27017` |
| Elasticsearch | `9200` / `9300` |
| Prometheus | `9090` |
| Zookeeper | `2181` |
| Kafka | `9092` / `9093` |
| Kafka UI | `18090` |

说明：

- 仓库中提供了 Prometheus 配置和 tracing 相关代码。
- `docker-compose-infra.yml` 目前没有 Jaeger、Zipkin、Pyroscope 容器，因此 README 不再把它们写成默认可直接启动的能力。

## Quick Start

### 1. Install Dependencies

```bash
go mod download
```

前端依赖会在执行 `make start-frontend` 时自动安装；如果你想手动安装：

```bash
cd frontend-user && npm install
cd frontend-admin && npm install
```

### 2. Start Infrastructure

```bash
make start-infra
```

检查容器状态：

```bash
docker compose -f docker-compose-infra.yml ps
```

### 3. Initialize MySQL

确保本地 MySQL 可用，并创建数据库：

```bash
mysql -uroot -p123456 -e "CREATE DATABASE IF NOT EXISTS ecommerce DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"
mysql -uroot -p123456 ecommerce < database/schema.sql
```

也可以直接使用初始化脚本：

```bash
bash database/init.sh
```

### 4. Check Development Configs

当前仓库使用的是 `configs/dev/*.yaml`。这些文件已经存在，但里面默认写的是本地开发地址，例如：

- MySQL: `localhost:3306`
- Redis: `localhost:6379`
- Kafka: `127.0.0.1:9092`
- Elasticsearch: `http://localhost:9200`

如果你的环境不同，先修改对应配置文件再启动。

### 5. Start Backend

一键启动后端与网关：

```bash
make start-backend
```

这个入口最终会调用 `scripts/start-all.sh --gateway`，会检查基础设施、数据库、配置文件，并启动核心服务、扩展服务、后台消费者和网关。

如果只想启动单个服务：

```bash
./scripts/start-service.sh user-service configs/dev/user-config.yaml
./scripts/start-service.sh product-service configs/dev/product-config.yaml
```

### 6. Start Frontend

```bash
make start-frontend
```

启动后访问：

- 用户端: [http://localhost:5173](http://localhost:5173)
- 管理端: [http://localhost:5174](http://localhost:5174)
- API Gateway: [http://localhost:8080](http://localhost:8080)
- Swagger UI: [http://localhost:8095](http://localhost:8095)

## Common Commands

```bash
make help
make build
make build-service SERVICE=user-service
make test
make lint
make clean
make proto
make proto-descriptor
make swagger
```

## Seckill Flow

秒杀链路由以下组件组成：

- `seckill-service`
- `order-service-consumer`
- Redis 秒杀库存键
- Kafka 异步消息

常用命令：

```bash
make seckill-init
make seckill-start
make seckill-check
make redis-set-stock SKU_ID=1 STOCK=100
make redis-get-stock SKU_ID=1
make redis-list-stocks
make seckill-stop
make seckill-full
```

处理流程：

```text
Client
  -> API Gateway
  -> Seckill Service
  -> Redis 原子扣减库存
  -> Kafka 投递秒杀订单消息
  -> Order Service Consumer 异步创建订单
```

## Frontend Notes

- 两个前端都使用 Vite
- 两个前端都通过代理访问 `http://localhost:8080`
- 用户端页面位于 `frontend-user/src/pages`
- 管理端页面位于 `frontend-admin/src/pages`

## API Documentation

Swagger 文件位于 `docs/swagger/`，网关启动后会在 `8095` 端口提供 Swagger UI。

常见接口前缀：

- `/api/v1/user/*`
- `/api/v1/products`
- `/api/v1/skus`
- `/api/v1/categories`
- `/api/v1/orders`
- `/api/v1/cart`
- `/api/v1/payment`
- `/api/v1/seckill/*`
- `/api/v1/search`
- `/api/v1/files/upload`

## Data Seeder

仓库包含一个 Python 爬虫，位于 `cmd/mi-crawler/`。使用方式见：

- [cmd/mi-crawler/README.md](/Users/sunweilin/Documents/Personal-Codebase/go-ecom/cmd/mi-crawler/README.md)

## Review Notes

这次 README 更新主要修正了以下与仓库实际状态不一致的地方：

- 把“16 个微服务”改成与当前仓库一致的服务描述，区分领域服务、gateway 和 consumer
- 删除了不存在的 `configs/test`、`configs/prod` 目录描述
- 删除了默认可用的 Jaeger、Zipkin、Pyroscope 基础设施表述
- 保留并强化了当前真实可用的启动命令、端口、配置文件和前端入口
- 明确了 `make start-backend`、`make start-frontend`、秒杀相关命令的实际用法

## Validation

已执行的检查：

- `go test ./...`
  代码测试主体通过；退出时因沙箱限制无法清理本机 Go build cache，出现 `failed to trim cache`，不属于业务代码失败。
- README 内容已逐项对照 `configs/dev/*.yaml`、`docker-compose-infra.yml`、`Makefile`、`scripts/start-all.sh`、`cmd/api-gateway/main.go` 进行校正。
