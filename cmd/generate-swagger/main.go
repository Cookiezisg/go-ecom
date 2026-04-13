package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// GatewayConfig Gateway 配置结构
type GatewayConfig struct {
	Upstreams []Upstream `yaml:"Upstreams"`
}

type Upstream struct {
	Name     string    `yaml:"Name"`
	Mappings []Mapping `yaml:"Mappings"`
}

type Mapping struct {
	Method  string `yaml:"Method"`
	Path    string `yaml:"Path"`
	RpcPath string `yaml:"RpcPath"`
}

// SwaggerDoc Swagger 文档结构
type SwaggerDoc struct {
	Swagger     string                            `json:"swagger"`
	Info        map[string]interface{}            `json:"info"`
	Host        string                            `json:"host,omitempty"`
	BasePath    string                            `json:"basePath,omitempty"`
	Schemes     []string                          `json:"schemes,omitempty"`
	Tags        []map[string]interface{}          `json:"tags"`
	Consumes    []string                          `json:"consumes"`
	Produces    []string                          `json:"produces"`
	Paths       map[string]map[string]interface{} `json:"paths"`
	Definitions map[string]interface{}            `json:"definitions"`
}

type ProtoField struct {
	Name     string
	Type     string
	Repeated bool
	IsMap    bool
	MapKey   string
	MapValue string
	Comment  string
}

type ProtoMessage struct {
	Name   string
	Fields []ProtoField
}

type ProtoEnum struct {
	Name string
}

func main() {
	// 读取 Gateway 配置
	gatewayConfig, err := loadGatewayConfig("configs/dev/gateway.yaml")
	if err != nil {
		log.Fatalf("加载 Gateway 配置失败: %v", err)
	}

	// 读取现有的 Swagger JSON 文件，收集所有处理后的文档用于合并
	swaggerDir := "docs/swagger"
	var allDocs []SwaggerDoc

	err = filepath.Walk(swaggerDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// 跳过合并输出文件自身，避免重复处理
		if !strings.HasSuffix(path, ".swagger.json") || strings.HasSuffix(path, "all.swagger.json") {
			return nil
		}

		// 读取现有 Swagger JSON
		data, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}

		var swagger SwaggerDoc
		if err := json.Unmarshal(data, &swagger); err != nil {
			return err
		}

		// 从路径推断服务名
		// 例如: docs/swagger/user/v1/user.swagger.json -> user
		parts := strings.Split(filepath.ToSlash(path), "/")
		var serviceName string
		for i, part := range parts {
			if part == "swagger" && i+1 < len(parts) {
				serviceName = parts[i+1]
				break
			}
		}

		// 查找对应的 Gateway 映射
		var upstream *Upstream
		for i := range gatewayConfig.Upstreams {
			up := &gatewayConfig.Upstreams[i]
			if strings.Contains(strings.ToLower(up.Name), serviceName) {
				upstream = up
				break
			}
		}

		if swagger.Definitions == nil {
			swagger.Definitions = make(map[string]interface{})
		}

		protoDefs, err := loadProtoDefinitions(serviceName)
		if err != nil {
			log.Printf("警告: 加载 %s 的 proto definitions 失败: %v", serviceName, err)
		} else {
			for name, def := range protoDefs {
				if _, exists := swagger.Definitions[name]; !exists {
					swagger.Definitions[name] = def
				}
			}
		}

		// 生成 paths
		if upstream != nil {
			swagger.Paths = generatePaths(upstream.Mappings, swagger.Definitions, serviceName)
		}

		swagger.Host = "localhost:8080"
		swagger.BasePath = ""
		swagger.Schemes = []string{"http"}

		// 保存更新后的单服务 Swagger JSON
		output, err := json.MarshalIndent(swagger, "", "  ")
		if err != nil {
			return err
		}
		if err := ioutil.WriteFile(path, output, 0644); err != nil {
			return err
		}

		fmt.Printf("✅ 已更新: %s\n", path)
		allDocs = append(allDocs, swagger)
		return nil
	})

	if err != nil {
		log.Fatalf("处理 Swagger 文件失败: %v", err)
	}

	// 合并所有服务的文档，输出 all.swagger.json
	if len(allDocs) > 0 {
		merged := mergeSwaggerDocs(allDocs)
		output, err := json.MarshalIndent(merged, "", "  ")
		if err != nil {
			log.Fatalf("合并 Swagger 文档失败: %v", err)
		}
		allPath := filepath.Join(swaggerDir, "all.swagger.json")
		if err := ioutil.WriteFile(allPath, output, 0644); err != nil {
			log.Fatalf("写入 all.swagger.json 失败: %v", err)
		}
		fmt.Printf("✅ 已生成合并文档: %s\n", allPath)
	}

	fmt.Println("\n✅ 所有 Swagger 文档已更新完成！")
}

func mergeSwaggerDocs(docs []SwaggerDoc) SwaggerDoc {
	merged := SwaggerDoc{
		Swagger: "2.0",
		Info: map[string]interface{}{
			"title":   "Go Ecom API",
			"version": "1.0",
		},
		Host:        "localhost:8080",
		BasePath:    "",
		Schemes:     []string{"http"},
		Consumes:    []string{"application/json"},
		Produces:    []string{"application/json"},
		Paths:       make(map[string]map[string]interface{}),
		Definitions: make(map[string]interface{}),
		Tags:        []map[string]interface{}{},
	}

	tagSet := make(map[string]bool)
	for _, doc := range docs {
		for path, methods := range doc.Paths {
			merged.Paths[path] = methods
		}
		for name, def := range doc.Definitions {
			merged.Definitions[name] = def
		}
	}

	// 从实际 paths 里重建 tags，避免 proto 生成的空 tag 混入
	for _, methods := range merged.Paths {
		for _, op := range methods {
			opMap, ok := op.(map[string]interface{})
			if !ok {
				continue
			}
			tags, ok := opMap["tags"].([]string)
			if !ok {
				continue
			}
			for _, t := range tags {
				if !tagSet[t] {
					tagSet[t] = true
					merged.Tags = append(merged.Tags, map[string]interface{}{"name": t})
				}
			}
		}
	}
	return merged
}

// rpcServiceName 从 RPC 路径提取服务名，例如 "user.v1.UserService/Register" -> "UserService"
func rpcServiceName(rpcPath string) string {
	parts := strings.Split(strings.Split(rpcPath, "/")[0], ".")
	return parts[len(parts)-1]
}

func loadGatewayConfig(path string) (*GatewayConfig, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config GatewayConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

func generatePaths(mappings []Mapping, definitions map[string]interface{}, serviceName string) map[string]map[string]interface{} {
	paths := make(map[string]map[string]interface{})

	for _, mapping := range mappings {
		// 将 Gateway 路径格式转换为 Swagger 格式
		// :id -> {id}
		path := convertPathParams(mapping.Path)
		method := strings.ToLower(mapping.Method)
		if method == "options" {
			continue
		}

		if paths[path] == nil {
			paths[path] = make(map[string]interface{})
		}

		// 从 RpcPath 提取请求和响应类型
		// 例如: cart.v1.CartService/GetCart -> GetCartRequest, GetCartResponse
		rpcName := strings.Split(mapping.RpcPath, "/")[1]
		requestType := fmt.Sprintf("v1%sRequest", rpcName)
		responseType := fmt.Sprintf("v1%sResponse", rpcName)

		// 构建操作对象
		operation := map[string]interface{}{
			"tags":        []string{rpcServiceName(mapping.RpcPath)},
			"summary":     getSummaryFromRpcName(rpcName),
			"operationId": rpcName,
			"consumes":    []string{"application/json"},
			"produces":    []string{"application/json"},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "成功",
					"schema": map[string]interface{}{
						"$ref": fmt.Sprintf("#/definitions/%s", responseType),
					},
				},
				"default": map[string]interface{}{
					"description": "错误",
					"schema": map[string]interface{}{
						"$ref": "#/definitions/rpcStatus",
					},
				},
			},
		}

		// 如果是 POST/PUT，添加请求体
		if method == "post" || method == "put" {
			// 检查 definitions 中是否有请求类型，如果没有或为空则创建/更新
			if existingDef, ok := definitions[requestType]; !ok {
				// 从 proto 文件生成请求类型定义
				definitions[requestType] = generateRequestDefinition(rpcName, serviceName)
			} else {
				// 如果定义存在但为空，则重新生成
				if defMap, ok := existingDef.(map[string]interface{}); ok {
					if props, ok := defMap["properties"].(map[string]interface{}); !ok || len(props) == 0 {
						definitions[requestType] = generateRequestDefinition(rpcName, serviceName)
					}
				}
			}
			operation["parameters"] = []map[string]interface{}{
				{
					"name":        "body",
					"in":          "body",
					"required":    true,
					"description": "请求体",
					"schema": map[string]interface{}{
						"$ref": fmt.Sprintf("#/definitions/%s", requestType),
					},
				},
			}
		} else {
			// GET/DELETE 方法，从路径提取参数
			params := extractPathParams(path)
			if len(params) > 0 {
				operation["parameters"] = params
			}
		}

		paths[path][method] = operation
	}

	return paths
}

func convertPathParams(path string) string {
	// 将 :param 转换为 {param}
	parts := strings.Split(path, "/")
	for i, part := range parts {
		if strings.HasPrefix(part, ":") {
			paramName := strings.TrimPrefix(part, ":")
			parts[i] = "{" + paramName + "}"
		}
	}
	return strings.Join(parts, "/")
}

func loadProtoDefinitions(serviceName string) (map[string]interface{}, error) {
	pattern := filepath.Join("api", serviceName, "*", serviceName+".proto")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}
	if len(matches) == 0 {
		return nil, fmt.Errorf("未找到 proto 文件: %s", pattern)
	}

	protoFile := matches[0]
	version := filepath.Base(filepath.Dir(protoFile))
	messages, enums, err := parseProtoFile(protoFile)
	if err != nil {
		return nil, err
	}

	definitions := make(map[string]interface{}, len(messages))
	for _, msg := range messages {
		definitions[version+msg.Name] = buildMessageDefinition(msg, version, messages, enums)
	}
	return definitions, nil
}

func parseProtoFile(path string) (map[string]ProtoMessage, map[string]ProtoEnum, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, nil, err
	}

	lines := strings.Split(string(data), "\n")
	messageStartRe := regexp.MustCompile(`^\s*message\s+([A-Za-z0-9_]+)\s*\{`)
	enumStartRe := regexp.MustCompile(`^\s*enum\s+([A-Za-z0-9_]+)\s*\{`)
	repeatedFieldRe := regexp.MustCompile(`^\s*repeated\s+([.\w]+)\s+([A-Za-z0-9_]+)\s*=\s*\d+`)
	mapFieldRe := regexp.MustCompile(`^\s*map<\s*([.\w]+)\s*,\s*([.\w]+)\s*>\s+([A-Za-z0-9_]+)\s*=\s*\d+`)
	fieldRe := regexp.MustCompile(`^\s*([.\w]+)\s+([A-Za-z0-9_]+)\s*=\s*\d+`)

	messages := make(map[string]ProtoMessage)
	enums := make(map[string]ProtoEnum)

	var currentMessage *ProtoMessage
	var currentEnum *ProtoEnum
	braceDepth := 0
	pendingComment := ""

	for _, rawLine := range lines {
		line := stripInlineComment(rawLine)
		trimmed := strings.TrimSpace(line)
		comment := strings.TrimSpace(extractInlineComment(rawLine))

		if trimmed == "" {
			if comment == "" {
				pendingComment = ""
			}
			continue
		}

		if strings.HasPrefix(trimmed, "//") {
			text := strings.TrimSpace(strings.TrimPrefix(trimmed, "//"))
			if pendingComment == "" {
				pendingComment = text
			} else {
				pendingComment += " " + text
			}
			continue
		}

		if currentMessage == nil && currentEnum == nil {
			if matches := messageStartRe.FindStringSubmatch(trimmed); len(matches) > 0 {
				currentMessage = &ProtoMessage{Name: matches[1]}
				braceDepth = 1
				pendingComment = ""
				continue
			}
			if matches := enumStartRe.FindStringSubmatch(trimmed); len(matches) > 0 {
				currentEnum = &ProtoEnum{Name: matches[1]}
				braceDepth = 1
				pendingComment = ""
				continue
			}
			pendingComment = ""
			continue
		}

		braceDepth += strings.Count(trimmed, "{")
		braceDepth -= strings.Count(trimmed, "}")

		if currentMessage != nil {
			switch {
			case repeatedFieldRe.MatchString(trimmed):
				matches := repeatedFieldRe.FindStringSubmatch(trimmed)
				currentMessage.Fields = append(currentMessage.Fields, ProtoField{
					Name:     matches[2],
					Type:     baseProtoType(matches[1]),
					Repeated: true,
					Comment:  fieldComment(pendingComment, comment),
				})
			case mapFieldRe.MatchString(trimmed):
				matches := mapFieldRe.FindStringSubmatch(trimmed)
				currentMessage.Fields = append(currentMessage.Fields, ProtoField{
					Name:     matches[3],
					IsMap:    true,
					MapKey:   baseProtoType(matches[1]),
					MapValue: baseProtoType(matches[2]),
					Comment:  fieldComment(pendingComment, comment),
				})
			case fieldRe.MatchString(trimmed) && !strings.HasPrefix(trimmed, "option ") && !strings.HasPrefix(trimmed, "reserved "):
				matches := fieldRe.FindStringSubmatch(trimmed)
				currentMessage.Fields = append(currentMessage.Fields, ProtoField{
					Name:    matches[2],
					Type:    baseProtoType(matches[1]),
					Comment: fieldComment(pendingComment, comment),
				})
			}
		}

		pendingComment = ""

		if braceDepth == 0 {
			if currentMessage != nil {
				messages[currentMessage.Name] = *currentMessage
				currentMessage = nil
			}
			if currentEnum != nil {
				enums[currentEnum.Name] = *currentEnum
				currentEnum = nil
			}
		}
	}

	return messages, enums, nil
}

func buildMessageDefinition(msg ProtoMessage, version string, messages map[string]ProtoMessage, enums map[string]ProtoEnum) map[string]interface{} {
	properties := make(map[string]interface{}, len(msg.Fields))
	for _, field := range msg.Fields {
		properties[field.Name] = buildFieldSchema(field, version, messages, enums)
	}
	return map[string]interface{}{
		"type":       "object",
		"properties": properties,
	}
}

func buildFieldSchema(field ProtoField, version string, messages map[string]ProtoMessage, enums map[string]ProtoEnum) map[string]interface{} {
	var schema map[string]interface{}

	switch {
	case field.IsMap:
		schema = map[string]interface{}{
			"type":                 "object",
			"additionalProperties": protoTypeSchema(field.MapValue, version, messages, enums),
		}
	case field.Repeated:
		schema = map[string]interface{}{
			"type":  "array",
			"items": protoTypeSchema(field.Type, version, messages, enums),
		}
	default:
		schema = protoTypeSchema(field.Type, version, messages, enums)
	}

	if field.Comment != "" {
		schema["description"] = field.Comment
	}
	return schema
}

func protoTypeSchema(fieldType, version string, messages map[string]ProtoMessage, enums map[string]ProtoEnum) map[string]interface{} {
	switch fieldType {
	case "string":
		return map[string]interface{}{"type": "string"}
	case "bool":
		return map[string]interface{}{"type": "boolean"}
	case "bytes":
		return map[string]interface{}{"type": "string", "format": "byte"}
	case "double":
		return map[string]interface{}{"type": "number", "format": "double"}
	case "float":
		return map[string]interface{}{"type": "number", "format": "float"}
	case "int32", "sint32", "sfixed32", "fixed32", "uint32":
		return map[string]interface{}{"type": "integer", "format": "int32"}
	case "int64", "sint64", "sfixed64", "fixed64", "uint64":
		return map[string]interface{}{"type": "string", "format": "int64"}
	}

	if _, ok := enums[fieldType]; ok {
		return map[string]interface{}{"type": "integer", "format": "int32"}
	}
	if _, ok := messages[fieldType]; ok {
		return map[string]interface{}{"$ref": fmt.Sprintf("#/definitions/%s%s", version, fieldType)}
	}
	return map[string]interface{}{"type": "string"}
}

func baseProtoType(fieldType string) string {
	parts := strings.Split(fieldType, ".")
	return parts[len(parts)-1]
}

func stripInlineComment(line string) string {
	if idx := strings.Index(line, "//"); idx >= 0 {
		return line[:idx]
	}
	return line
}

func extractInlineComment(line string) string {
	if idx := strings.Index(line, "//"); idx >= 0 {
		return strings.TrimSpace(line[idx+2:])
	}
	return ""
}

func fieldComment(pendingComment, inlineComment string) string {
	if inlineComment != "" {
		return inlineComment
	}
	return pendingComment
}

func extractPathParams(path string) []map[string]interface{} {
	var params []map[string]interface{}
	parts := strings.Split(path, "/")

	for _, part := range parts {
		// 支持两种格式：:param 和 {param}
		var paramName string
		if strings.HasPrefix(part, ":") {
			paramName = strings.TrimPrefix(part, ":")
		} else if strings.HasPrefix(part, "{") && strings.HasSuffix(part, "}") {
			paramName = strings.TrimPrefix(strings.TrimSuffix(part, "}"), "{")
		}

		if paramName != "" {
			params = append(params, map[string]interface{}{
				"name":        paramName,
				"in":          "path",
				"required":    true,
				"type":        "string",
				"description": fmt.Sprintf("%s 参数", paramName),
			})
		}
	}

	return params
}

func generateRequestDefinition(rpcName, serviceName string) map[string]interface{} {
	// 根据 RPC 方法名生成请求类型定义
	// 这里定义常见请求类型的字段
	requestFields := map[string]map[string]interface{}{
		"Register": {
			"username": map[string]interface{}{
				"type":        "string",
				"description": "用户名",
			},
			"password": map[string]interface{}{
				"type":        "string",
				"description": "密码",
			},
			"phone": map[string]interface{}{
				"type":        "string",
				"description": "手机号",
			},
			"email": map[string]interface{}{
				"type":        "string",
				"description": "邮箱",
			},
			"verifyCode": map[string]interface{}{
				"type":        "string",
				"description": "验证码",
			},
		},
		"Login": {
			"username": map[string]interface{}{
				"type":        "string",
				"description": "用户名/手机号/邮箱",
			},
			"password": map[string]interface{}{
				"type":        "string",
				"description": "密码",
			},
			"loginType": map[string]interface{}{
				"type":        "integer",
				"format":      "int32",
				"description": "登录类型: 1-用户名, 2-手机号, 3-邮箱",
			},
		},
		"AddAddress": {
			"userId": map[string]interface{}{
				"type":        "string",
				"format":      "int64",
				"description": "用户ID",
			},
			"receiverName": map[string]interface{}{
				"type":        "string",
				"description": "收货人姓名",
			},
			"receiverPhone": map[string]interface{}{
				"type":        "string",
				"description": "收货人电话",
			},
			"province": map[string]interface{}{
				"type":        "string",
				"description": "省份",
			},
			"city": map[string]interface{}{
				"type":        "string",
				"description": "城市",
			},
			"district": map[string]interface{}{
				"type":        "string",
				"description": "区县",
			},
			"detail": map[string]interface{}{
				"type":        "string",
				"description": "详细地址",
			},
			"postalCode": map[string]interface{}{
				"type":        "string",
				"description": "邮编",
			},
			"isDefault": map[string]interface{}{
				"type":        "integer",
				"format":      "int32",
				"description": "是否默认地址",
			},
		},
		"UpdateAddress": {
			"id": map[string]interface{}{
				"type":        "string",
				"format":      "int64",
				"description": "地址ID",
			},
			"userId": map[string]interface{}{
				"type":        "string",
				"format":      "int64",
				"description": "用户ID",
			},
			"receiverName": map[string]interface{}{
				"type":        "string",
				"description": "收货人姓名",
			},
			"receiverPhone": map[string]interface{}{
				"type":        "string",
				"description": "收货人电话",
			},
			"province": map[string]interface{}{
				"type":        "string",
				"description": "省份",
			},
			"city": map[string]interface{}{
				"type":        "string",
				"description": "城市",
			},
			"district": map[string]interface{}{
				"type":        "string",
				"description": "区县",
			},
			"detail": map[string]interface{}{
				"type":        "string",
				"description": "详细地址",
			},
			"postalCode": map[string]interface{}{
				"type":        "string",
				"description": "邮编",
			},
			"isDefault": map[string]interface{}{
				"type":        "integer",
				"format":      "int32",
				"description": "是否默认地址",
			},
		},
		"UpdateUserInfo": {
			"userId": map[string]interface{}{
				"type":        "string",
				"format":      "int64",
				"description": "用户ID",
			},
			"nickname": map[string]interface{}{
				"type":        "string",
				"description": "昵称",
			},
			"avatar": map[string]interface{}{
				"type":        "string",
				"description": "头像",
			},
			"gender": map[string]interface{}{
				"type":        "integer",
				"format":      "int32",
				"description": "性别",
			},
			"birthday": map[string]interface{}{
				"type":        "string",
				"description": "生日",
			},
		},
		"AddItem": {
			"userId": map[string]interface{}{
				"type":        "string",
				"format":      "int64",
				"description": "用户ID",
			},
			"skuId": map[string]interface{}{
				"type":        "string",
				"format":      "int64",
				"description": "SKU ID",
			},
			"quantity": map[string]interface{}{
				"type":        "integer",
				"format":      "int32",
				"description": "数量",
			},
		},
		"UpdateQuantity": {
			"userId": map[string]interface{}{
				"type":        "string",
				"format":      "int64",
				"description": "用户ID",
			},
			"skuId": map[string]interface{}{
				"type":        "string",
				"format":      "int64",
				"description": "SKU ID",
			},
			"quantity": map[string]interface{}{
				"type":        "integer",
				"format":      "int32",
				"description": "数量",
			},
		},
		"CancelOrder": {
			"id": map[string]interface{}{
				"type":        "string",
				"format":      "int64",
				"description": "订单ID",
			},
			"orderNo": map[string]interface{}{
				"type":        "string",
				"description": "订单号",
			},
			"reason": map[string]interface{}{
				"type":        "string",
				"description": "取消原因",
			},
		},
		"Refund": {
			"paymentNo": map[string]interface{}{
				"type":        "string",
				"description": "支付单号",
			},
			"refundAmount": map[string]interface{}{
				"type":        "string",
				"description": "退款金额",
			},
			"reason": map[string]interface{}{
				"type":        "string",
				"description": "退款原因",
			},
		},
		"CreateOrder": {
			"userId": map[string]interface{}{
				"type":        "string",
				"format":      "int64",
				"description": "用户ID",
			},
			"addressId": map[string]interface{}{
				"type":        "string",
				"format":      "int64",
				"description": "地址ID",
			},
			"items": map[string]interface{}{
				"type":        "array",
				"description": "订单项",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"skuId": map[string]interface{}{
							"type":   "string",
							"format": "int64",
						},
						"quantity": map[string]interface{}{
							"type":   "integer",
							"format": "int32",
						},
					},
				},
			},
		},
		"CreatePayment": {
			"orderId": map[string]interface{}{
				"type":        "string",
				"format":      "int64",
				"description": "订单ID",
			},
			"paymentMethod": map[string]interface{}{
				"type":        "string",
				"description": "支付方式",
			},
			"amount": map[string]interface{}{
				"type":        "number",
				"format":      "double",
				"description": "支付金额",
			},
		},
		"DeductStock": {
			"skuId": map[string]interface{}{
				"type":        "string",
				"format":      "int64",
				"description": "SKU ID",
			},
			"quantity": map[string]interface{}{
				"type":        "integer",
				"format":      "int32",
				"description": "扣减数量",
			},
		},
	}

	if fields, ok := requestFields[rpcName]; ok {
		return map[string]interface{}{
			"type":       "object",
			"properties": fields,
		}
	}

	// 默认返回空对象
	return map[string]interface{}{
		"type":       "object",
		"properties": map[string]interface{}{},
	}
}

func getSummaryFromRpcName(rpcName string) string {
	// 简单的转换，将驼峰命名转换为中文描述
	descriptions := map[string]string{
		"GetCart":         "获取购物车",
		"AddItem":         "添加商品到购物车",
		"UpdateQuantity":  "更新商品数量",
		"RemoveItem":      "删除商品",
		"ClearCart":       "清空购物车",
		"SelectItem":      "选择商品",
		"BatchSelect":     "批量选择",
		"ListProducts":    "获取商品列表",
		"GetProduct":      "获取商品详情",
		"GetSku":          "获取SKU详情",
		"GetCategoryList": "获取分类列表",
		"GetCategoryTree": "获取分类树",
		"Register":        "用户注册",
		"Login":           "用户登录",
		"GetUserInfo":     "获取用户信息",
		"UpdateUserInfo":  "更新用户信息",
		"GetAddressList":  "获取地址列表",
		"AddAddress":      "添加地址",
		"UpdateAddress":   "更新地址",
		"DeleteAddress":   "删除地址",
		"CreateOrder":     "创建订单",
		"GetOrder":        "获取订单详情",
		"ListOrders":      "获取订单列表",
		"CancelOrder":     "取消订单",
		"CreatePayment":   "创建支付",
		"GetPayment":      "获取支付信息",
		"Refund":          "退款",
		"GetInventory":    "获取库存",
		"DeductStock":     "扣减库存",
	}

	if desc, ok := descriptions[rpcName]; ok {
		return desc
	}
	return rpcName
}
