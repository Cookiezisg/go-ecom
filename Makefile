.PHONY: build test lint clean proto proto-descriptor swagger api help start-all start-infra stop-infra seckill-init seckill-start seckill-stop seckill-full run-seckill run-order-consumer redis-cli redis-set-stock redis-get-stock redis-list-stocks seckill-check

# 项目名称
PROJECT_NAME := ecommerce-system

# Go 参数
GOCMD := go
GOBUILD := $(GOCMD) build
GOTEST := $(GOCMD) test
GOGET := $(GOCMD) get
GOMOD := $(GOCMD) mod

# 服务列表
SERVICES := user-service product-service order-service payment-service inventory-service cart-service seckill-service

help: ## Show help
	@echo "Available make targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \\033[36m%-15s\\033[0m %s\\n", $$1, $$2}'

build: ## Build all services
	@echo "Building all services..."
ifeq ($(OS),Windows_NT)
	@for %%s in ($(SERVICES)) do ( \\
		echo Building %%s... & \\
		$(GOBUILD) -o bin/%%s.exe ./cmd/%%s \\
	)
else
	@for service in $(SERVICES); do \\
		echo "Building $$service..."; \\
		$(GOBUILD) -o bin/$$service ./cmd/$$service; \\
	done
endif
	@echo "Build finished."

build-service: ## Build single service (usage: make build-service SERVICE=user-service)
	@if [ -z "$(SERVICE)" ]; then \\
		echo "Error: please specify SERVICE, e.g. make build-service SERVICE=user-service"; \\
		exit 1; \\
	fi
	@echo "Building $(SERVICE)..."
	$(GOBUILD) -o bin/$(SERVICE) ./cmd/$(SERVICE)
	@echo "Build finished."

test: ## Run tests
	@echo "Running tests..."
	$(GOTEST) -v ./... -cover
	@echo "Tests finished."

lint: ## Run linters
	@echo "Running linters..."
	@if command -v golangci-lint > /dev/null; then \\
		golangci-lint run; \\
	else \\
		echo "Warning: golangci-lint not installed, skip lint step"; \\
	fi

clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	rm -rf dist/
	@echo "Clean finished."

proto: ## 生成 Protobuf 代码
ifeq ($(OS),Windows_NT)
	@powershell -NoProfile -ExecutionPolicy Bypass -File scripts\\generate-proto.ps1
else
	@if command -v protoc > /dev/null; then \\
		find api -name "*.proto" -exec protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative {} \\; ; \\
	else \\
		echo "错误: 未安装 protoc，请先安装 Protocol Buffers"; \\
		exit 1; \\
	fi
endif

proto-descriptor: ## 生成 Proto Descriptor 文件（用于 Gateway）
	@echo "生成 Proto Descriptor 文件..."
	@if [ -f scripts/generate-proto-descriptor.sh ]; then \\
		chmod +x scripts/generate-proto-descriptor.sh && \\
		./scripts/generate-proto-descriptor.sh; \\
	else \\
		echo "错误: 脚本文件不存在"; \\
		exit 1; \\
	fi
	@echo "Proto Descriptor 文件生成完成！"

swagger: ## 从 .proto 生成 OpenAPI (Swagger) 文档到 docs/swagger
ifeq ($(OS),Windows_NT)
	@echo "在 Windows 上生成 OpenAPI (Swagger) 文档..."
	@if exist scripts\\generate-openapi.bat ( \\
		scripts\\generate-openapi.bat \\
	) else ( \\
		echo 错误: scripts\\generate-openapi.bat 不存在; \\
		exit 1; \\
	)
else
	@echo "在 Unix/Linux/Mac 上生成 OpenAPI (Swagger) 文档..."
	@if [ -f scripts/generate-openapi.sh ]; then \\
		chmod +x scripts/generate-openapi.sh && \\
		./scripts/generate-openapi.sh; \\
	else \\
		echo "错误: scripts/generate-openapi.sh 不存在"; \\
		exit 1; \\
	fi
endif

api: ## 使用 goctl 生成 API 代码 (需要先安装 goctl)
	@echo "使用 goctl 生成 API 代码..."
	@if command -v goctl > /dev/null; then \\
		echo "goctl 已安装"; \\
	else \\
		echo "错误: 未安装 goctl，请运行: go install github.com/zeromicro/go-zero/tools/goctl@latest"; \\
		exit 1; \\
	fi

deps: ## 下载依赖
	@echo "下载依赖..."
	$(GOMOD) download
	$(GOMOD) tidy
	@echo "依赖下载完成！"

init: ## Initialize project structure (create directories)
	@echo "Initializing project structure..."
	@mkdir -p api/user/v1 api/product/v1 api/order/v1
	@mkdir -p cmd/user-service cmd/product-service cmd/order-service cmd/seckill-service cmd/order-service-consumer
	@mkdir -p internal/pkg/{cache,database,logger,middleware}
	@mkdir -p internal/service/{user,product,order}
	@mkdir -p pkg/{utils,constants,errors}
	@mkdir -p configs/{dev,test,prod}
	@mkdir -p deployments/{docker,k8s,scripts}
	@mkdir -p docs test bin
	@echo "Project structure initialized."

run-user: ## Run user service (dev)
	@echo "Running user-service..."
ifeq ($(OS),Windows_NT)
	$(GOBUILD) -o bin/user-service.exe ./cmd/user-service
	.\\bin\\user-service.exe
else
	$(GOBUILD) -o bin/user-service ./cmd/user-service && ./bin/user-service
endif

run-product: ## Run product service (dev)
	@echo "Running product-service..."
ifeq ($(OS),Windows_NT)
	$(GOBUILD) -o bin/product-service.exe ./cmd/product-service
	.\\bin\\product-service.exe
else
	$(GOBUILD) -o bin/product-service ./cmd/product-service && ./bin/product-service
endif

run-seckill: ## Run seckill service (dev)
	@echo "Running seckill-service..."
ifeq ($(OS),Windows_NT)
	$(GOBUILD) -o bin/seckill-service.exe ./cmd/seckill-service
	.\\bin\\seckill-service.exe -f configs/dev/seckill-config.yaml
else
	$(GOBUILD) -o bin/seckill-service ./cmd/seckill-service && ./bin/seckill-service -f configs/dev/seckill-config.yaml
endif

run-order-consumer: ## Run order service consumer (dev)
	@echo "Running order-service-consumer..."
ifeq ($(OS),Windows_NT)
	$(GOBUILD) -o bin/order-service-consumer.exe ./cmd/order-service-consumer
	.\\bin\\order-service-consumer.exe -f configs/dev/order-config.yaml
else
	$(GOBUILD) -o bin/order-service-consumer ./cmd/order-service-consumer && ./bin/order-service-consumer -f configs/dev/order-config.yaml
endif

start-all: ## Start all services (infrastructure + microservices)
	@echo "Starting all services..."
ifeq ($(OS),Windows_NT)
	@powershell -NoProfile -ExecutionPolicy Bypass -File scripts\\start-all.ps1
else
	@chmod +x scripts/start-all.sh && ./scripts/start-all.sh --gateway
endif

start-services: ## Start microservices only (skip infrastructure)
	@echo "Starting microservices (skipping infrastructure)..."
ifeq ($(OS),Windows_NT)
	@powershell -NoProfile -ExecutionPolicy Bypass -File scripts\\start-all.ps1 --skip-infra
else
	@chmod +x scripts/start-all.sh && ./scripts/start-all.sh --gateway --skip-infra
endif

start-infra: ## Start infrastructure services (Docker Compose)
	@echo "Starting infrastructure services..."
	@docker compose -f docker-compose-infra.yml down --remove-orphans 2>/dev/null || true
	@docker rm -f infra-redis infra-etcd infra-etcd-keeper infra-mongodb infra-elasticsearch infra-prometheus 2>/dev/null || true
	@docker compose -f docker-compose-infra.yml up -d
	@echo "Infrastructure services started. Use 'make stop-infra' to stop them."

stop-infra: ## Stop infrastructure services
	@echo "Stopping infrastructure services..."
	@docker compose -f docker-compose-infra.yml down
	@echo "Infrastructure services stopped."

seckill-init: ## Initialize seckill (generate proto, setup redis stock)
	@echo "初始化秒杀功能..."
	@echo "1. 生成 Proto 文件..."
	@protoc --go_out=. --go-grpc_out=. api/seckill/v1/seckill.proto 2>/dev/null || echo "警告: proto 文件可能已存在或 protoc 未安装"
	@echo "2. 初始化 Redis 库存（示例）..."
	@echo "   请手动执行以下命令设置库存："
	@echo "   make redis-set-stock SKU_ID=1 STOCK=100"
	@echo "   make redis-set-stock SKU_ID=2 STOCK=50"
	@echo "   或者使用 docker exec:"
	@echo "   docker exec -it infra-redis redis-cli SET seckill:stock:1 100"
	@echo "初始化完成！"

redis-cli: ## Connect to Redis CLI in Docker
	@echo "连接到 Docker 中的 Redis..."
	@docker exec -it infra-redis redis-cli

redis-set-stock: ## Set seckill stock in Redis (usage: make redis-set-stock SKU_ID=1 STOCK=100)
	@if [ -z "$(SKU_ID)" ] || [ -z "$(STOCK)" ]; then \\
		echo "错误: 请指定 SKU_ID 和 STOCK"; \\
		echo "用法: make redis-set-stock SKU_ID=1 STOCK=100"; \\
		exit 1; \\
	fi
	@echo "设置秒杀库存: SKU_ID=$(SKU_ID), STOCK=$(STOCK)"
	@docker exec infra-redis redis-cli SET "seckill:stock:$(SKU_ID)" "$(STOCK)" || \\
		(echo "错误: 无法连接到 Redis，请确保 Docker 容器正在运行: make start-infra" && exit 1)
	@echo "✓ 库存设置成功"

redis-get-stock: ## Get seckill stock from Redis (usage: make redis-get-stock SKU_ID=1)
	@if [ -z "$(SKU_ID)" ]; then \\
		echo "错误: 请指定 SKU_ID"; \\
		echo "用法: make redis-get-stock SKU_ID=1"; \\
		exit 1; \\
	fi
	@echo "查询秒杀库存: SKU_ID=$(SKU_ID)"
	@docker exec infra-redis redis-cli GET "seckill:stock:$(SKU_ID)" || \\
		(echo "错误: 无法连接到 Redis，请确保 Docker 容器正在运行: make start-infra" && exit 1)

redis-list-stocks: ## List all seckill stocks
	@echo "查询所有秒杀库存..."
	@docker exec infra-redis redis-cli KEYS "seckill:stock:*" | while read key; do \\
		if [ -n "$$key" ]; then \\
			value=$$(docker exec infra-redis redis-cli GET "$$key"); \\
			echo "  $$key = $$value"; \\
		fi \\
	done || echo "错误: 无法连接到 Redis，请确保 Docker 容器正在运行: make start-infra"

seckill-check: ## Check seckill service status and dependencies
	@echo "检查秒杀服务状态..."
	@chmod +x scripts/check-seckill.sh && ./scripts/check-seckill.sh

seckill-start: ## Start seckill service and consumer (requires infra running)
	@echo "启动秒杀服务..."
	@echo "确保基础设施服务已启动（Redis、Kafka）: make start-infra"
	@echo ""
	@echo "启动秒杀服务（后台运行）..."
ifeq ($(OS),Windows_NT)
	@start /B $(GOBUILD) -o bin/seckill-service.exe ./cmd/seckill-service && start /B .\\bin\\seckill-service.exe -f configs/dev/seckill-config.yaml
	@timeout /t 2 /nobreak >nul
	@start /B $(GOBUILD) -o bin/order-service-consumer.exe ./cmd/order-service-consumer && start /B .\\bin\\order-service-consumer.exe -f configs/dev/order-config.yaml
else
	@mkdir -p logs
	@$(GOBUILD) -o bin/seckill-service ./cmd/seckill-service
	@nohup ./bin/seckill-service -f configs/dev/seckill-config.yaml > logs/seckill-service.log 2>&1 & \\
	echo $$! > /tmp/seckill-service.pid && \\
	echo "秒杀服务已启动 (PID: $$(cat /tmp/seckill-service.pid))"
	@sleep 2
	@$(GOBUILD) -o bin/order-service-consumer ./cmd/order-service-consumer
	@nohup ./bin/order-service-consumer -f configs/dev/order-config.yaml > logs/order-consumer.log 2>&1 & \\
	echo $$! > /tmp/order-consumer.pid && \\
	echo "订单服务消费者已启动 (PID: $$(cat /tmp/order-consumer.pid))"
	@echo ""
	@echo "服务已启动，日志文件："
	@echo "  - logs/seckill-service.log"
	@echo "  - logs/order-consumer.log"
	@echo ""
	@echo "停止服务：make seckill-stop"
endif

seckill-stop: ## Stop seckill service and consumer
	@echo "停止秒杀服务..."
ifeq ($(OS),Windows_NT)
	@taskkill /F /IM seckill-service.exe 2>nul || echo "秒杀服务未运行"
	@taskkill /F /IM order-service-consumer.exe 2>nul || echo "订单消费者未运行"
else
	@if [ -f /tmp/seckill-service.pid ]; then \\
		kill $$(cat /tmp/seckill-service.pid) 2>/dev/null && rm /tmp/seckill-service.pid || echo "秒杀服务未运行"; \\
	else \\
		pkill -f seckill-service || echo "秒杀服务未运行"; \\
	fi
	@if [ -f /tmp/order-consumer.pid ]; then \\
		kill $$(cat /tmp/order-consumer.pid) 2>/dev/null && rm /tmp/order-consumer.pid || echo "订单消费者未运行"; \\
	else \\
		pkill -f order-service-consumer || echo "订单消费者未运行"; \\
	fi
endif
	@echo "秒杀服务已停止"

seckill-full: start-infra seckill-init seckill-start ## Full seckill setup (infra + init + start)
	@echo ""
	@echo "=========================================="
	@echo "秒杀功能已完全启动！"
	@echo "=========================================="
	@echo "服务地址："
	@echo "  - 秒杀服务: localhost:8090"
	@echo "  - Kafka UI: http://localhost:18090 (如果已启动)"
	@echo ""
	@echo "测试命令："
	@echo "  grpcurl -plaintext -d '{\\"user_id\\": 1, \\"sku_id\\": 1, \\"quantity\\": 1}' localhost:8090 seckill.v1.SeckillService/Seckill"
	@echo ""
	@echo "查看日志："
	@echo "  tail -f logs/seckill-service.log"
	@echo "  tail -f logs/order-consumer.log"
	@echo ""

.DEFAULT_GOAL := help



