# 小米商城爬虫 (Python版本)

使用 Python + Selenium + BeautifulSoup4 + requests 实现的小米商城爬虫。

## 功能特点

- ✅ 使用 Selenium 等待页面完全加载，确保获取到动态内容
- ✅ 使用 BeautifulSoup4 解析 HTML
- ✅ 使用 requests 下载图片
- ✅ 从商品详情页提取原图（b2c-shopapi-pms 格式）
- ✅ 支持分类和商品的爬取
- ✅ 自动保存到 MySQL 数据库

## 安装依赖

默认直接从官方 PyPI 安装，不再默认使用国内镜像源。

```bash
pip install -r requirements.txt
```

## 安装 ChromeDriver

### macOS
```bash
brew install chromedriver
```

### Linux
```bash
# Ubuntu/Debian
sudo apt-get install chromium-chromedriver

# 或下载对应版本的 chromedriver
wget https://chromedriver.storage.googleapis.com/LATEST_RELEASE/chromedriver_linux64.zip
unzip chromedriver_linux64.zip
sudo mv chromedriver /usr/local/bin/
```

### Windows
下载对应版本的 ChromeDriver: https://chromedriver.chromium.org/

## 使用方法

```bash
# 基本用法
python main.py

# 指定数据库配置
python main.py \
  --db-host localhost \
  --db-port 3306 \
  --db-user root \
  --db-password 123456 \
  --db-name ecommerce \
  --image-dir images/mi
```

## 三阶段流程（按你的要求：分类 → 商品 → SKU）

默认启动会先做清理（**删除图片 + 清空数据库**），然后按阶段执行：

- 1) **分类 + 分类图入库**
- 2) **商品 + 商品图入库**（旧 DOM 下目前仍在分类阶段内顺带爬商品；新 DOM 需要你确认子分类结构后再完全拆出）
- 3) **SKU 入库**（从商品详情页尽量提取“选择版本/选择颜色”，生成组合写入 `sku` 表）

```bash
# 默认：清库 + 删图 + 跑 categories,products,skus
python main.py

# 只跑分类（仍会清库删图，除非 --no-reset）
python main.py --stages categories

# SKU 阶段只处理前 N 个商品（调试用）
python main.py --stages skus --sku-limit 5

# 控制数量：每个主分类最多 5 个子分类；每个子分类最多 50 个商品（默认就是这个）
python main.py --max-subcategories-per-parent 5 --max-products-per-subcategory 50
```

## 工作流程

1. **清空数据表**: 默认清空 `sku` / `product` / `category`
2. **删除旧图片**: 默认删除 `images/mi/category` 和 `images/mi/product`
3. **清理 Redis 缓存**: 默认清理 `category:tree:status:*` 等缓存，避免“清库后前端仍只显示旧分类树”
3. **爬取分类**: 
   - 访问小米商城首页
   - 提取主分类
   - 保存主分类到数据库
   - 爬取每个主分类的子分类
4. **爬取商品**:
   - 访问分类/搜索页面
   - 提取商品链接
   - 访问每个商品的详情页
   - 等待页面完全加载（使用 Selenium）
   - 提取商品信息和图片
   - 下载图片到本地
   - 保存到数据库
5. **爬取 SKU**:
   - 遍历 `product` 表
   - 进入详情页提取“选择版本/选择颜色”等选项
   - 生成规格组合并写入 `sku` 表（重复运行会更新并自动恢复软删除的同 sku_code）

## 图片提取策略

1. **优先使用 CSS 选择器**: 查找 `.swiper-slide img` 和 `.swiper-wrapper img`
2. **备用方案**: 如果 CSS 选择器没找到，使用正则表达式直接从 HTML 中提取
3. **图片URL处理**: 自动移除URL参数，确保下载原图
4. **文件大小检查**: 下载后显示文件大小，便于确认是否为原图

## 注意事项

- 需要安装 Chrome 浏览器和 ChromeDriver
- 默认使用无头模式（headless），可以在代码中修改
- 每个分类最多爬取5个商品，可以在代码中调整
- 图片下载有延迟，避免请求过快

## 与 Go 版本的区别

- ✅ 使用 Selenium 可以执行 JavaScript，确保页面完全加载
- ✅ 可以等待动态内容加载完成
- ✅ 更适合处理 Vue.js 等前端框架渲染的页面
- ⚠️ 运行速度较慢（需要启动浏览器）
- ⚠️ 资源消耗较大（需要运行 Chrome）
