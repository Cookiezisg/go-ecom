#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
小米商城爬虫
使用 Selenium + BeautifulSoup4 + requests 实现
"""

import os
import sys
import time
import json
import re
import argparse
import hashlib
from pathlib import Path
from urllib.parse import urljoin, urlparse
from typing import List, Dict, Optional

import requests
from bs4 import BeautifulSoup
from selenium import webdriver
from selenium.webdriver.common.by import By
from selenium.webdriver.support.ui import WebDriverWait
from selenium.webdriver.support import expected_conditions as EC
from selenium.webdriver.chrome.options import Options
from selenium.common.exceptions import TimeoutException, WebDriverException
from selenium.webdriver.common.action_chains import ActionChains
import pymysql
from pymysql.cursors import DictCursor
try:
    import redis  # type: ignore
except Exception:
    redis = None


class MiCrawler:
    def __init__(self, db_config: Dict, image_dir: str = "images/mi"):
        """
        初始化爬虫
        
        Args:
            db_config: 数据库配置
            image_dir: 图片保存目录
        """
        self.db_config = db_config
        
        # 确保图片目录是绝对路径，相对于项目根目录
        # 如果 image_dir 是相对路径，转换为相对于 ecommerce-backend 的绝对路径
        if not os.path.isabs(image_dir):
            # 查找 ecommerce-backend 目录
            current = Path.cwd()
            backend_dir = None
            for parent in [current] + list(current.parents):
                if parent.name == 'ecommerce-backend':
                    backend_dir = parent
                    break
            
            if backend_dir:
                self.image_dir = backend_dir / image_dir
            else:
                # 如果找不到，使用当前目录
                self.image_dir = Path(image_dir).resolve()
        else:
            self.image_dir = Path(image_dir)
        
        # 尝试创建目录，如果失败则使用当前目录
        try:
            self.image_dir.mkdir(parents=True, exist_ok=True)
        except (PermissionError, OSError) as e:
            print(f"警告: 无法在 {self.image_dir} 创建目录: {e}")
            # 使用当前工作目录下的images目录
            self.image_dir = Path.cwd() / image_dir
            try:
                self.image_dir.mkdir(parents=True, exist_ok=True)
                print(f"使用当前目录下的图片目录: {self.image_dir}")
            except Exception as e2:
                print(f"错误: 无法创建图片目录: {e2}")
                sys.exit(1)
        
        # 创建图片子目录
        (self.image_dir / "category").mkdir(exist_ok=True)
        (self.image_dir / "product").mkdir(exist_ok=True)
        
        print(f"图片保存目录: {self.image_dir}")
        
        # 初始化数据库连接
        self.db = None
        self.connect_db()
        
        # 初始化 Selenium WebDriver
        self.driver = None
        self.init_driver()
        
        # 初始化 requests session
        self.session = requests.Session()
        self.session.headers.update({
            'User-Agent': 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36',
            'Accept': 'image/webp,image/apng,image/*,*/*;q=0.8',
            'Accept-Language': 'zh-CN,zh;q=0.9,en;q=0.8',
            'Referer': 'https://www.mi.com/',
        })

        # 爬取限制（默认按你的要求）
        # - 每个主分类最多 5 个子分类
        # - 每个子分类最多 50 个商品（从二级列表页抓链接）
        self.max_subcategories_per_parent = 5
        self.max_products_per_subcategory = 50

        # Redis（用于清理分类树缓存，避免“清库后分类树还停留旧数据”）
        self.redis_host = self.db_config.get("redis_host", "localhost")
        self.redis_port = int(self.db_config.get("redis_port", 6379))
        self.redis_password = self.db_config.get("redis_password", "")
        self.redis_db = int(self.db_config.get("redis_db", 0))
    
    def connect_db(self):
        """连接数据库"""
        try:
            self.db = pymysql.connect(
                host=self.db_config['host'],
                port=self.db_config['port'],
                user=self.db_config['user'],
                password=self.db_config['password'],
                database=self.db_config['database'],
                charset='utf8mb4',
                cursorclass=DictCursor
            )
            print("数据库连接成功")
        except Exception as e:
            print(f"数据库连接失败: {e}")
            sys.exit(1)

    def reset_database_and_images(self, truncate_tables: bool = True, delete_images: bool = True):
        """
        爬取前重置环境：
        - 删除图片目录（images/mi 下的 category/product）
        - 清空数据库（category/product/sku）
        - 清理 Redis 缓存（category tree 等）
        """
        import shutil

        if delete_images:
            print("\n========== 清理图片目录 ==========")
            for sub in ["category", "product"]:
                d = self.image_dir / sub
                try:
                    if d.exists():
                        shutil.rmtree(d)
                    d.mkdir(parents=True, exist_ok=True)
                    print(f"  ✓ 已清理: {d}")
                except Exception as e:
                    print(f"  ⚠️  清理失败 {d}: {e}")

        if truncate_tables:
            print("\n========== 清空数据库表 ==========")
            with self.db.cursor() as cursor:
                # 避免外键约束导致 TRUNCATE 失败
                cursor.execute("SET FOREIGN_KEY_CHECKS=0")
                # 注意顺序：先子表再父表
                cursor.execute("TRUNCATE TABLE sku")
                cursor.execute("TRUNCATE TABLE product")
                cursor.execute("TRUNCATE TABLE category")
                cursor.execute("SET FOREIGN_KEY_CHECKS=1")
            self.db.commit()
            print("  ✓ 已清空: sku / product / category")

        # 清理 Redis 缓存（非常关键：后端 GetCategoryTree 会缓存 24h，清库不清缓存会导致前端仍只看到旧的分类树）
        self.reset_redis_cache()

    def reset_redis_cache(self):
        """清理后端 Redis 缓存（分类树为主，顺带清理部分商品相关缓存）"""
        print("\n========== 清理 Redis 缓存 ==========")
        if redis is None:
            print("  ⚠️  未安装 redis 依赖，跳过缓存清理（请先 pip install -r requirements.txt）")
            return
        try:
            r = redis.Redis(
                host=self.redis_host,
                port=self.redis_port,
                password=self.redis_password or None,
                db=self.redis_db,
                decode_responses=True,
                socket_connect_timeout=2,
                socket_timeout=2,
            )
            # ping 一下确保可用
            r.ping()

            # 关键：分类树缓存 key（按 status 分开）
            keys = [
                "category:tree:status:-1",
                "category:tree:status:0",
                "category:tree:status:1",
                "category:tree:status:2",
            ]
            deleted = 0
            for k in keys:
                deleted += int(r.delete(k) or 0)

            # 保险：如果存在其它历史 key 形式，按前缀扫描删除（数量通常很少）
            scan_deleted = 0
            for pattern in ["category:tree*", "product:list:*", "product:detail:*", "sku:info:*"]:
                for k in r.scan_iter(match=pattern, count=200):
                    scan_deleted += int(r.delete(k) or 0)

            print(f"  ✓ 已删除分类树缓存key: {deleted} 个；扫描删除: {scan_deleted} 个")
        except Exception as e:
            print(f"  ⚠️  Redis 缓存清理失败（不影响爬虫入库，但可能影响前端展示）: {e}")
    
    def init_driver(self):
        """初始化 Selenium WebDriver"""
        chrome_options = Options()
        chrome_options.add_argument('--headless')  # 无头模式
        chrome_options.add_argument('--no-sandbox')
        chrome_options.add_argument('--disable-dev-shm-usage')
        chrome_options.add_argument('--disable-gpu')
        chrome_options.add_argument('--window-size=1920,1080')
        chrome_options.add_argument('user-agent=Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36')
        
        try:
            self.driver = webdriver.Chrome(options=chrome_options)
            self.driver.implicitly_wait(10)
            print("WebDriver 初始化成功")
        except Exception as e:
            print(f"WebDriver 初始化失败: {e}")
            print("请确保已安装 Chrome 和 chromedriver")
            sys.exit(1)
    
    def wait_for_page_load(self, timeout=30, wait_for_images=True):
        """等待页面加载完成"""
        try:
            # 1. 等待页面 DOM 加载完成
            WebDriverWait(self.driver, timeout).until(
                lambda driver: driver.execute_script('return document.readyState') == 'complete'
            )
            print("      ✓ DOM 加载完成")
            
            if wait_for_images:
                # 2. 等待图片元素出现（关键：等待真实图片src替换完成）
                WebDriverWait(self.driver, timeout).until(
                    EC.presence_of_element_located((By.CSS_SELECTOR, "img"))
                )
                print("      ✓ 图片元素已出现")
                
                # 3. 额外等待图片懒加载替换（关键：JS会把placeholder替换成真实图片）
                # 小米商城是懒加载+JS渲染，必须等待JS执行完成
                time.sleep(3)  # 等待懒加载完成
                print("      ✓ 等待懒加载完成（3秒）")
                
                # 4. 等待真实图片加载完成（确保是真实图片，不是占位图）
                WebDriverWait(self.driver, timeout).until(
                    lambda driver: driver.execute_script(
                        """
                        var imgs = document.querySelectorAll('img');
                        var loaded = 0;
                        for (var i = 0; i < imgs.length; i++) {
                            var img = imgs[i];
                            if (img.complete && img.naturalWidth > 0) {
                                // 检查是否是真实图片（不是占位图）
                                var src = img.src || img.getAttribute('data-src') || img.getAttribute('data-original') || '';
                                if (src && !src.includes('placeholder') && !src.includes('lazy') && 
                                    (src.includes('b2c-shopapi-pms') || src.includes('cdn.cnbj') || src.includes('mi-img.com'))) {
                                    loaded++;
                                }
                            }
                        }
                        return loaded > 0;
                        """
                    )
                )
                print("      ✓ 真实图片已加载完成")
                
                # 5. 等待特定的商品图片元素出现（详情页）
                try:
                    WebDriverWait(self.driver, 5).until(
                        EC.presence_of_element_located((By.CSS_SELECTOR, ".swiper-slide img, .swiper-wrapper img, img[src*='b2c-shopapi-pms']"))
                    )
                    print("      ✓ 商品图片元素已加载")
                except TimeoutException:
                    print("      ⚠️  未找到商品图片元素，继续尝试提取...")
                except Exception as e:
                    print(f"      ⚠️  等待商品图片元素时出错: {e}")
                    try:
                        WebDriverWait(self.driver, 5).until(
                            EC.presence_of_element_located((By.CSS_SELECTOR, "img[src*='b2c-shopapi-pms']"))
                        )
                        print("      ✓ 商品图片URL已加载")
                    except TimeoutException:
                        print("      ⚠️  未找到商品图片元素，继续尝试提取...")
            
            return True
        except TimeoutException as e:
            print(f"      ⚠️  页面加载超时: {e}")
            return False
        except Exception as e:
            print(f"      ⚠️  等待页面加载时出错: {e}")
            return False
    
    def crawl_categories(self, base_url: str = "https://www.mi.com/shop", only_main_categories: bool = False):
        """
        爬取分类
        
        Args:
            base_url: 小米商城首页URL
            only_main_categories: 如果为True，只爬主分类，不爬子分类和商品
        """
        print("\n========== 开始爬取分类 ==========")
        
        if only_main_categories:
            print("【模式：只爬主分类】")
        else:
            print("【模式：爬取完整分类和商品】")
        
        # 注意：清库/删图已挪到 reset_database_and_images()，这里不再做副作用操作
        
        try:
            print(f"访问小米商城首页: {base_url}")
            self.driver.get(base_url)
            
            # 等待页面加载完成（首页需要等待图片，因为分类有图片）
            self.wait_for_page_load(timeout=20, wait_for_images=True)
            
            # 获取页面源码
            html = self.driver.page_source
            # 用 lxml 解析更稳（部分页面结构 html.parser 会漏节点，导致抓不到 swiper-wrapper）
            soup = BeautifulSoup(html, 'lxml')
            
            # 兼容两种分类 DOM：
            # 1) 旧版：li.category-item / a.title（你之前提供的 J_categoryList）
            # 2) 新版：div.category-nav / .category-nav-item / .category-nav-name（你这次贴的 DOM）
            parent_items = soup.select('li.category-item')
            if parent_items:
                print(f"\n找到 {len(parent_items)} 个主分类项（旧 DOM: li.category-item）")
            else:
                parent_names = [e.get_text(strip=True) for e in soup.select('.category-nav .category-nav-item .category-nav-name')]
                parent_names = [n for n in parent_names if n]
                print(f"\n找到 {len(parent_names)} 个主分类项（新 DOM: .category-nav）")

                # 只入库主分类（子分类需要 hover 才出现，后续再补充）
                sort_order = 0
                for name in parent_names:
                    print(f"\n========== [{sort_order + 1}] 处理主分类: {name} ==========")
                    parent_id = self.save_category_with_image(name, 0, 1, "")
                    if parent_id <= 0:
                        print("  ⚠️  主分类保存失败")
                    else:
                        print(f"  ✓ 主分类已保存，ID: {parent_id}")
                    sort_order += 1
                print(f"\n========== 分类爬取完成 ==========")
                print(f"共爬取 {sort_order} 个主分类")
                return
            
            sort_order = 0
            for parent_item in parent_items:
                # 提取主分类名称
                title_elem = parent_item.select_one('a.title')
                if not title_elem:
                    continue
                
                parent_name = title_elem.get_text(strip=True)
                if not parent_name:
                    continue
                
                # 清理名称（移除可能的箭头图标文本）
                parent_name = re.sub(r'\s+', ' ', parent_name).strip()
                # 移除箭头图标文本（如果有）
                parent_name = re.sub(r'<em.*?</em>', '', parent_name, flags=re.DOTALL)
                parent_name = parent_name.strip()
                
                # 提取主分类图片
                # 根据test.html结构，主分类本身没有图片，但实际页面可能通过JS动态加载
                # 尝试从主分类项中查找图片
                parent_image_url = ""
                
                # 方法1: 查找主分类项内的第一个有效图片（排除子分类的图片）
                # 排除子分类区域的图片（div.children 内的图片）
                main_category_imgs = parent_item.select('img')
                for img in main_category_imgs:
                    # 检查是否在子分类区域内
                    parent_div = img.find_parent('div', class_='children')
                    if parent_div:
                        continue  # 跳过子分类的图片
                    
                    # 优先级：data-original > data-lazy-src > data-src > src
                    img_url = (img.get('data-original') or 
                              img.get('data-lazy-src') or 
                              img.get('data-src') or 
                              img.get('src') or "")
                    
                    if img_url:
                        # 过滤占位图和箭头图标
                        img_lower = img_url.lower()
                        is_placeholder = any(kw in img_lower for kw in ['placeholder', 'default', 'empty', 'blank', 'arrow', 'iconfont'])
                        
                        if not is_placeholder:
                            # 规范化URL
                            if img_url.startswith('//'):
                                img_url = 'https:' + img_url
                            elif img_url.startswith('/'):
                                img_url = 'https://www.mi.com' + img_url
                            elif not img_url.startswith('http'):
                                img_url = urljoin('https://www.mi.com', img_url)
                            
                            # 移除URL参数
                            img_url = re.sub(r'[?&](thumb|w|h|f|q)=\d+', '', img_url)
                            img_url = img_url.rstrip('?&')
                            
                            # 只保留有效的CDN图片
                            if any(cdn in img_url for cdn in ['b2c-shopapi-pms', 'cdn.cnbj', 'mi-mall', 'nr-pub', 'mi-img.com', 'pms_']):
                                parent_image_url = img_url
                                break
                
                print(f"\n========== [{sort_order + 1}] 处理主分类: {parent_name} ==========")
                if parent_image_url:
                    print(f"  找到主分类图片: {parent_image_url}")
                else:
                    print(f"  未找到主分类图片（主分类可能没有图片）")
                
                # 保存主分类，获取ID（如果有图片则下载）
                parent_id = self.save_category_with_image(parent_name, 0, 1, parent_image_url)
                if parent_id <= 0:
                    print(f"  ⚠️  主分类保存失败，跳过")
                    continue
                
                print(f"  ✓ 主分类已保存，ID: {parent_id}")
                
                # 如果只爬主分类，跳过子分类和商品处理
                if only_main_categories:
                    sort_order += 1
                    continue
                
                # 提取子分类（从 div.children li 中提取）
                child_items = parent_item.select('div.children li')
                child_sort_order = 0
                
                # 子分类最多 5 个（按你的要求）
                for child_item in child_items[: self.max_subcategories_per_parent]:
                    # 提取子分类名称
                    text_elem = child_item.select_one('span.text')
                    if not text_elem:
                        continue
                    
                    child_name = text_elem.get_text(strip=True)
                    if not child_name:
                        continue
                    
                    # 提取子分类图片（从 img.thumb 中提取）
                    img_elem = child_item.select_one('img.thumb')
                    image_url = ""
                    if img_elem:
                        # 优先级：data-original > data-lazy-src > data-src > src
                        image_url = (img_elem.get('data-original') or 
                                    img_elem.get('data-lazy-src') or 
                                    img_elem.get('data-src') or 
                                    img_elem.get('src') or "")
                    
                    # 规范化图片URL
                    if image_url:
                        # 过滤占位图
                        img_lower = image_url.lower()
                        is_placeholder = any(kw in img_lower for kw in ['placeholder', 'default', 'empty', 'blank'])
                        
                        if not is_placeholder:
                            if image_url.startswith('//'):
                                image_url = 'https:' + image_url
                            elif image_url.startswith('/'):
                                image_url = 'https://www.mi.com' + image_url
                            elif not image_url.startswith('http'):
                                image_url = urljoin('https://www.mi.com', image_url)
                            
                            # 移除URL参数中的尺寸限制（参考Go版本）
                            if '?' in image_url or '&' in image_url:
                                image_url = re.sub(r'[?&](thumb|w|h|f|q)=\d+', '', image_url)
                                image_url = image_url.rstrip('?&')
                                # 如果还有?，直接移除所有参数获取原图
                                if '?' in image_url:
                                    image_url = image_url.split('?')[0]
                            
                            print(f"  提取子分类图片URL: {image_url}")
                    
                    # 提取子分类链接
                    link_elem = child_item.select_one('a.link')
                    child_url = ""
                    product_urls = []  # 存储产品链接
                    
                    if link_elem:
                        href = link_elem.get('href', '')
                        if href:
                            if href.startswith('//'):
                                child_url = 'https:' + href
                            elif href.startswith('/'):
                                child_url = 'https://www.mi.com' + href
                            elif href.startswith('http'):
                                child_url = href
                            else:
                                child_url = urljoin('https://www.mi.com', href)
                            
                            # 分类.html：子分类 href 可能是
                            # - /shop/buy?product_id=... （可直接转详情页）
                            # - /shop/search?keyword=... （进入二级产品列表页，只抓链接）
                            # 如果链接是产品链接（包含 product_id），直接添加到产品链接列表
                            if 'shop/buy?product_id' in child_url or 'product_id=' in child_url:
                                # 规范化产品链接为详情页格式
                                match = re.search(r'product_id=(\d+)', child_url)
                                if match:
                                    product_url = f"https://www.mi.com/shop/buy/detail?product_id={match.group(1)}"
                                    product_urls.append(product_url)
                                    print(f"  找到产品链接: {product_url}")
                            # 如果是搜索链接（/shop/search?keyword=），留给二级列表页抓商品链接
                            elif 'shop/search' in child_url or 'search?keyword' in child_url:
                                print(f"  子分类为搜索列表页: {child_url}")
                    
                    print(f"  处理子分类: {child_name} (图片: {image_url}, 产品链接数: {len(product_urls)})")
                    
                    # 保存子分类（使用已保存的主分类ID），获取子分类ID
                    child_id = self.save_category_with_image(child_name, parent_id, 2, image_url)
                    child_sort_order += 1
                    
                    # 如果子分类保存成功：
                    # - 若直接拿到详情页链接：直接爬详情（产品详情页.html）
                    # - 若是搜索列表页：进入二级产品列表（只取链接），再爬详情
                    if child_id > 0:
                        if product_urls:
                            # 有直接的产品链接，直接爬取
                            print(f"    开始爬取子分类 {child_name} (ID: {child_id}) 下的 {len(product_urls)} 个产品...")
                            for product_url in product_urls:
                                self.crawl_product_detail(child_id, product_url)
                            print(f"    子分类 {child_name} 的商品爬取完成")
                        elif child_url:
                            print(f"    开始爬取子分类 {child_name} (ID: {child_id}) 下的商品链接...")
                            self.crawl_products_by_category(child_id, child_url)
                            print(f"    子分类 {child_name} 的商品爬取完成")
                        else:
                            print(f"    子分类 {child_name} 没有产品链接，跳过")
                
                sort_order += 1
            
            print(f"\n========== 爬取完成 ==========")
            print(f"共爬取 {sort_order} 个主分类")
        
        except Exception as e:
            print(f"爬取分类失败: {e}")
            import traceback
            traceback.print_exc()
    
    
    def crawl_products_by_category(self, category_id: int, category_url: str):
        """二级产品列表页（如 二级产品列表.html）：只提取商品链接，再进入详情页爬数据"""
        print(f"    爬取分类ID {category_id} 的商品: {category_url}")
        
        try:
            # 如果URL是商品详情页，直接爬取
            if 'shop/buy' in category_url and 'product_id' in category_url:
                # 转换为详情页URL格式
                match = re.search(r'product_id=(\d+)', category_url)
                if match:
                    detail_url = f"https://www.mi.com/shop/buy/detail?product_id={match.group(1)}"
                    print(f"    直接爬取商品详情页: {detail_url}")
                    self.crawl_product_detail(category_id, detail_url)
                    return
            
            # 访问分类/搜索页面，只提取商品链接
            print(f"    访问分类页面: {category_url}")
            self.driver.get(category_url)
            
            # 等待页面加载完成（列表页不需要等待图片）
            self.wait_for_page_load(timeout=20, wait_for_images=False)

            # 二级产品列表是 JS 渲染的（readyState=complete 不代表商品卡片已出来）
            # 显式等待商品卡片出现，再抓链接
            try:
                WebDriverWait(self.driver, 20).until(
                    EC.presence_of_element_located(
                        (By.CSS_SELECTOR, ".goods-list-box .goods-item, .goods-list .goods-item")
                    )
                )
                # 给一点时间让 href 填充完整
                time.sleep(0.8)
            except TimeoutException:
                print("    ⚠️  等待商品卡片超时，继续尝试从当前页面提取链接...")
            
            html = self.driver.page_source
            soup = BeautifulSoup(html, 'lxml')
            
            # 查找商品链接（严格按“二级产品列表.html”的商品卡片 DOM）
            # 只从 .goods-list .goods-item > a[href*='product_id'] 抓，避免把顶部导航/分类菜单里的 product_id 也抓进来导致分类错乱
            product_links = []
            card_links = soup.select(".goods-list-box .goods-item > a[href*='product_id']")
            if not card_links:
                card_links = soup.select(".goods-list .goods-item > a[href*='product_id']")
            if not card_links:
                # 部分页面可能是 li 结构，轻量兜底，但仍限制在 goods-list 区域
                card_links = soup.select(".goods-list-box a[href*='product_id']")
            if not card_links:
                card_links = soup.select(".goods-list a[href*='product_id']")

            for link in card_links:
                href = link.get('href', '') or ''
                if 'product_id' not in href:
                    continue
                full_url = urljoin("https://www.mi.com", href)
                # 转换为详情页URL格式
                match = re.search(r'product_id=(\d+)', full_url)
                if match:
                    detail_url = f"https://www.mi.com/shop/buy/detail?product_id={match.group(1)}"
                    if detail_url not in product_links:
                        product_links.append(detail_url)
            
            print(f"    找到 {len(product_links)} 个商品链接")
            
            if len(product_links) == 0:
                print(f"    ⚠️  未找到商品链接，跳过")
                return
            
            # 每个子分类最多 50 个商品（按你的要求）
            product_links = product_links[: self.max_products_per_subcategory]
            max_products = len(product_links)
            for i, product_url in enumerate(product_links):
                print(f"    [{i+1}/{max_products}] 爬取商品详情: {product_url}")
                try:
                    self.crawl_product_detail(category_id, product_url)
                    # 不需要固定延迟，页面加载等待已经足够
                except Exception as e:
                    print(f"    ⚠️  爬取商品失败: {e}")
                    import traceback
                    traceback.print_exc()
                    continue
        
        except Exception as e:
            print(f"    爬取商品失败: {e}")
            import traceback
            traceback.print_exc()
    
    def crawl_product_detail(self, category_id: int, url: str):
        """爬取商品详情页"""
        try:
            print(f"      访问商品详情页: {url}")
            # 使用 Selenium 访问详情页，等待页面完全加载
            self.driver.get(url)
            # 等待页面完全加载（包括图片）
            print(f"      等待页面完全加载...")
            loaded = self.wait_for_page_load(timeout=30, wait_for_images=True)
            
            if not loaded:
                print("      ⚠️  页面可能未完全加载，继续尝试提取...")
            
            # 获取页面源码
            print(f"      提取页面内容...")
            html = self.driver.page_source
            
            # 检查是否包含商品图片
            if 'b2c-shopapi-pms' not in html and 'swiper-slide' not in html:
                print(f"      ⚠️  页面可能未完全加载，未找到图片元素")
            
            soup = BeautifulSoup(html, 'lxml')
            
            # 提取商品信息
            product = {
                'url': url,
                'category_id': category_id,
                'name': '',
                'price': '',
                'images': [],
                'brand': '小米',
                'description': ''
            }
            
            # 提取商品名称
            # 优先按 产品详情页.html 的结构：.product-con h2
            h2 = soup.select_one('.product-con h2')
            if h2:
                product['name'] = h2.get_text(" ", strip=True)

            # 仅在没有 h2 时再用 title 兜底（避免把“立即购买-小米商城”等拼进 name）
            if not product['name']:
                title_tag = soup.find('title')
                if title_tag:
                    title = title_tag.get_text(strip=True)
                    product['name'] = title.split(' - ')[0] if ' - ' in title else title
            
            if not product['name']:
                h1 = soup.find('h1')
                if h1:
                    product['name'] = h1.get_text(strip=True)

            # 描述（产品详情页.html：p.sale-desc）
            desc_el = soup.select_one('.sale-desc')
            if desc_el:
                product['description'] = desc_el.get_text(" ", strip=True)
            
            # 提取价格
            # 优先按 产品详情页.html 的结构：.price-info
            price_elem = soup.select_one('.price-info')
            if price_elem:
                price_text = price_elem.get_text(" ", strip=True)
                match = re.search(r'[\d,]+\.?\d*', price_text)
                if match:
                    product['price'] = match.group(0).replace(',', '')
            else:
                price_selectors = ['.price', '.mi-price', '[class*="price"]']
                for selector in price_selectors:
                    pe = soup.select_one(selector)
                    if pe:
                        pt = pe.get_text(strip=True)
                        m = re.search(r'[\d,]+\.?\d*', pt)
                        if m:
                            product['price'] = m.group(0).replace(',', '')
                            break
            
            # 提取产品图：只爬你给的轮播区域（产品详情页.html）
            # 目标 DOM：.J_slider_imglist / .product-imglist 下的 swiper-wrapper -> swiper-slide -> img
            # 注意：部分商品页可能没有显式 .swiper-wrapper，但仍有 .swiper-slide img；此时退到同一轮播区域内的 slide img
            print("      提取图片（仅轮播区域 J_slider_imglist/swiper）...")

            image_map: Dict[str, bool] = {}

            # 1) 优先按文档结构：J_slider_imglist / product-imglist 下找 wrapper
            wrapper_candidates = [
                soup.select_one(".J_slider_imglist .swiper-wrapper"),
                soup.select_one(".product-imglist .swiper-wrapper"),
                soup.select_one(".J_slider_imglist [class*='swiper-wrapper']"),
                soup.select_one(".product-imglist [class*='swiper-wrapper']"),
            ]
            wrapper = next((w for w in wrapper_candidates if w is not None), None)

            # 2) 如果找不到 wrapper，就锁定轮播容器（避免误抓其它图片）
            slider_root = soup.select_one(".J_slider_imglist") or soup.select_one(".product-imglist")

            def _push_img_url(u: str):
                if len(product["images"]) >= 20:
                    return
                u = (u or "").strip()
                if not u:
                    return
                if u.startswith("//"):
                    u = "https:" + u
                # 只要轮播原图域名
                if "b2c-shopapi-pms" not in u:
                    return
                if u in image_map:
                    return
                image_map[u] = True
                product["images"].append(u)
                print(f"        提取到轮播原图 [{len(product['images'])}]: {u}")

            if wrapper:
                for img in wrapper.select("img[src]"):
                    _push_img_url(img.get("src"))
            elif slider_root:
                # 退到轮播区域内的 slide img（仍然只在轮播区域内）
                slide_imgs = slider_root.select(".swiper-slide img[src]") or slider_root.select("img[src]")
                for img in slide_imgs:
                    _push_img_url(img.get("src"))
                if len(product["images"]) == 0:
                    print("      ⚠️  轮播区域存在，但未匹配到 b2c-shopapi-pms 图片")
            else:
                print("      ⚠️  未找到轮播区域（J_slider_imglist/product-imglist），产品图为空")
            
            print(f"      共提取到 {len(product['images'])} 张图片")
            
            # 保存商品
            if product['name']:
                product_id = self.save_product(product)
                # 详情页同步入库 SKU（版本/颜色）
                if product_id:
                    try:
                        # 读取已入库的 spu_code / 图片，用于 sku_code 和 sku.image
                        with self.db.cursor() as cursor:
                            cursor.execute(
                                "SELECT spu_code, name, price, local_main_image, main_image FROM product WHERE id = %s",
                                (int(product_id),)
                            )
                            row = cursor.fetchone() or {}

                        spu_code = row.get("spu_code") or f"P{int(product_id)}"
                        pname = row.get("name") or product["name"]
                        base_price = float(row.get("price") or (float(product["price"]) if product.get("price") else 0))
                        sku_image = row.get("local_main_image") or row.get("main_image") or ""
                        if not sku_image and product.get("images"):
                            sku_image = product["images"][0]

                        opts = self.extract_sku_options(soup)
                        versions = opts.get("选择版本") or ["默认"]
                        colors = opts.get("选择颜色") or ["默认"]

                        # 默认库存：页面没给分 SKU 库存时，给一个可购买值
                        stock = 5

                        total = 0
                        for v in versions:
                            for c in colors:
                                specs: Dict[str, str] = {}
                                if v != "默认":
                                    specs["选择版本"] = v
                                if c != "默认":
                                    specs["选择颜色"] = c
                                spec_key = json.dumps(specs, ensure_ascii=False, sort_keys=True)
                                h = hashlib.md5(spec_key.encode("utf-8")).hexdigest()[:8]
                                sku_code = f"{spu_code}-{h}"
                                sku_name = f"{pname} {v} {c}".strip()
                                self.upsert_sku(
                                    product_id=int(product_id),
                                    sku_code=sku_code,
                                    name=sku_name,
                                    specs=specs,
                                    price=base_price,
                                    stock=stock,
                                    image=sku_image,
                                    status=1,
                                )
                                total += 1

                        print(f"      ✓ SKU入库完成：{total} 条（版本={len(versions)} 颜色={len(colors)}）")
                    except Exception as e:
                        print(f"      ⚠️  SKU提取/入库失败: {e}")
            else:
                print(f"      ⚠️  商品名称为空，跳过保存")
        
        except Exception as e:
            print(f"      爬取商品详情失败: {e}")
            import traceback
            traceback.print_exc()
    
    def save_category(self, name: str, parent_id: int, level: int) -> int:
        """保存分类到数据库（无图片）"""
        return self.save_category_with_image(name, parent_id, level, "")
    
    def save_category_with_image(self, name: str, parent_id: int, level: int, image_url: str = "") -> int:
        """保存分类到数据库（带图片）"""
        with self.db.cursor() as cursor:
            # 检查是否已存在
            cursor.execute(
                "SELECT id FROM category WHERE name = %s AND parent_id = %s AND level = %s",
                (name, parent_id, level)
            )
            existing = cursor.fetchone()
            
            if existing:
                category_id = existing['id']
                # 如果已有分类但图片为空，尝试更新图片
                if image_url:
                    cursor.execute(
                        "UPDATE category SET image = %s WHERE id = %s",
                        (image_url, category_id)
                    )
                    self.db.commit()
                return category_id
            
            # 下载分类图片（参考Go版本逻辑）
            local_image = ""
            if image_url:
                # 检查是否是有效的图片URL（CDN域名）
                is_valid_image = any(cdn in image_url for cdn in ['b2c-shopapi-pms', 'cdn.cnbj', 'mi-mall', 'nr-pub', 'mi-img.com', 'pms_'])
                if is_valid_image:
                    try:
                        print(f"    下载分类图片: {image_url}")
                        local_image = self.download_image(image_url, self.image_dir / "category")
                        if local_image:
                            print(f"    分类图片下载成功: {local_image}")
                    except Exception as e:
                        print(f"    分类图片下载失败: {e}")
                else:
                    print(f"    跳过无效的分类图片URL: {image_url}")
            
            # 插入新分类
            cursor.execute(
                """INSERT INTO category (name, parent_id, level, image, image_local, status, created_at, updated_at)
                   VALUES (%s, %s, %s, %s, %s, 1, NOW(), NOW())""",
                (name, parent_id, level, image_url if image_url else None, local_image if local_image else None)
            )
            self.db.commit()
            return cursor.lastrowid
    
    def save_product(self, product: Dict) -> Optional[int]:
        """保存商品到数据库，返回 product.id"""
        with self.db.cursor() as cursor:
            # 检查是否已存在
            cursor.execute(
                "SELECT id FROM product WHERE source_url = %s",
                (product['url'],)
            )
            existing = cursor.fetchone()
            if existing:
                print(f"      商品已存在，跳过: {product['name']}")
                return int(existing["id"])
            
            # 下载图片
            local_images = []
            local_main_image = ""
            main_image_url = ""
            
            if product['images']:
                print(f"      开始下载 {len(product['images'])} 张图片...")
                for i, img_url in enumerate(product['images'][:20]):  # 最多20张
                    try:
                        local_path = self.download_image(img_url, self.image_dir / "product")
                        if local_path:
                            local_images.append(local_path)
                            if i == 0:
                                local_main_image = local_path
                                main_image_url = img_url
                            print(f"        下载成功 [{i+1}/{len(product['images'])}]: {local_path}")
                        # 不需要固定延迟，图片下载本身有网络延迟
                    except Exception as e:
                        print(f"        下载失败: {e}")
            
            # 生成SPU编码
            product_id_match = re.search(r'product_id=(\d+)', product['url'])
            product_id = product_id_match.group(1) if product_id_match else str(hash(product['url']))
            spu_code = f"MI_{product_id}"
            
            # 检查SPU编码是否已存在
            cursor.execute("SELECT id FROM product WHERE spu_code = %s", (spu_code,))
            if cursor.fetchone():
                spu_code = f"{spu_code}_{int(time.time())}"
            
            # 插入商品
            cursor.execute(
                """INSERT INTO product 
                   (spu_code, name, subtitle, category_id, main_image, local_main_image,
                    images, local_images, price, original_price, source_url, brand_name,
                    crawled_at, status)
                   VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, NOW(), %s)""",
                (
                    spu_code,
                    product['name'],
                    product.get('description', ''),
                    product['category_id'],
                    main_image_url,
                    local_main_image,
                    json.dumps(product['images'], ensure_ascii=False),
                    json.dumps(local_images, ensure_ascii=False),
                    float(product['price']) if product['price'] else 0,
                    0,
                    product['url'],
                    product.get('brand', '小米'),
                    2  # 待审核
                )
            )
            self.db.commit()
            print(f"      保存商品成功: {product['name']}")
            return int(cursor.lastrowid)

    def extract_sku_options(self, soup: BeautifulSoup) -> Dict[str, List[str]]:
        """
        从详情页 DOM 提取 SKU 规格选项（尽量通用）
        目标：抓到类似“选择版本/选择颜色”的选项列表
        """
        options: Dict[str, List[str]] = {}

        # 优先按 产品详情页.html 的结构解析：
        # <div class="buy-option"><div class="option-box"><div class="title">选择版本</div><ul><li title="..."><a>...</a>
        buy_boxes = soup.select(".buy-option .option-box")
        for box in buy_boxes:
            title_el = box.select_one(".title")
            if not title_el:
                continue
            title = title_el.get_text(strip=True)
            if not title:
                continue
            vals: List[str] = []
            for li in box.select("ul li"):
                v = (li.get("title") or li.get_text(strip=True) or "").strip()
                if not v:
                    continue
                if v not in vals:
                    vals.append(v)
            if vals:
                options[title] = vals

        if options:
            return options

        text = soup.get_text("\n", strip=True)
        if not text:
            return options

        # 经验：小米详情页常见的 label 文案
        candidates = ["选择版本", "选择颜色", "版本", "颜色"]

        for label in candidates:
            # 找到包含 label 的节点
            label_nodes = soup.find_all(string=re.compile(re.escape(label)))
            for ln in label_nodes:
                parent = ln.parent
                if not parent:
                    continue
                # 尝试在 label 节点后面的同级/父级区域找按钮/选项文本
                container = parent.parent or parent
                # 选项可能是 li / a / span / div 按钮
                raw_opts = []
                for sel in ["li", "a", "span", "div"]:
                    for el in container.find_all(sel):
                        t = el.get_text(strip=True)
                        if not t:
                            continue
                        # 过滤掉 label 本身、价格、库存等噪声
                        if label in t:
                            continue
                        if any(k in t for k in ["¥", "库存", "件", "已选", "请选择"]):
                            continue
                        # 太长的说明文字不要
                        if len(t) > 40:
                            continue
                        raw_opts.append(t)
                # 去重并保序
                dedup = []
                for o in raw_opts:
                    if o not in dedup:
                        dedup.append(o)
                if dedup:
                    options[label] = dedup
                    break
            if label in options:
                continue

        # 归一化 key：优先输出“选择版本/选择颜色”
        if "版本" in options and "选择版本" not in options:
            options["选择版本"] = options.pop("版本")
        if "颜色" in options and "选择颜色" not in options:
            options["选择颜色"] = options.pop("颜色")

        return options

    def upsert_sku(self, product_id: int, sku_code: str, name: str, specs: Dict[str, str],
                   price: float, stock: int, image: str, status: int = 1):
        """
        写入 SKU（可重复运行）：ON DUPLICATE KEY UPDATE + 恢复 deleted_at
        """
        specs_json = json.dumps(specs or {}, ensure_ascii=False)
        with self.db.cursor() as cursor:
            cursor.execute(
                """INSERT INTO sku
                   (product_id, sku_code, name, specs, price, original_price, stock, image, status, created_at, updated_at, deleted_at)
                   VALUES (%s, %s, %s, %s, %s, NULL, %s, %s, %s, NOW(), NOW(), NULL)
                   ON DUPLICATE KEY UPDATE
                     deleted_at = NULL,
                     product_id = VALUES(product_id),
                     name = VALUES(name),
                     specs = VALUES(specs),
                     price = VALUES(price),
                     stock = VALUES(stock),
                     image = VALUES(image),
                     status = VALUES(status),
                     updated_at = NOW()
                """,
                (product_id, sku_code, name, specs_json, float(price), int(stock), image, int(status))
            )
        self.db.commit()

    def crawl_skus_for_products(self, limit: int = 0):
        """
        第三阶段：遍历数据库中的商品，进入详情页提取 SKU，并写入 sku 表
        """
        print("\n========== 开始爬取 SKU ==========")
        with self.db.cursor() as cursor:
            sql = "SELECT id, name, spu_code, price, local_main_image, main_image, source_url FROM product ORDER BY id ASC"
            if limit and limit > 0:
                sql += f" LIMIT {int(limit)}"
            cursor.execute(sql)
            products = cursor.fetchall()

        print(f"共 {len(products)} 个商品需要爬取 SKU")

        for idx, p in enumerate(products, start=1):
            pid = int(p["id"])
            pname = p.get("name") or ""
            spu_code = p.get("spu_code") or f"P{pid}"
            url = p.get("source_url") or ""
            base_price = float(p.get("price") or 0)
            image = p.get("local_main_image") or p.get("main_image") or ""

            if not url:
                continue

            print(f"\n[{idx}/{len(products)}] SKU: {pname} (product_id={pid})")
            try:
                self.driver.get(url)
                self.wait_for_page_load(timeout=30, wait_for_images=True)
                html = self.driver.page_source
                soup = BeautifulSoup(html, "html.parser")

                opts = self.extract_sku_options(soup)
                versions = opts.get("选择版本") or ["默认"]
                colors = opts.get("选择颜色") or ["默认"]

                # 详情页通常会显示一个“库存：X 件”，但它未必是分 SKU 的；这里先用它做默认库存
                stock = 0
                stock_match = re.search(r"库存[:：]?\s*(\d+)\s*件", soup.get_text(" ", strip=True))
                if stock_match:
                    stock = int(stock_match.group(1))
                if stock <= 0:
                    stock = 5  # 默认给一个可购买库存

                total = 0
                for v in versions:
                    for c in colors:
                        specs = {}
                        if v != "默认":
                            specs["选择版本"] = v
                        if c != "默认":
                            specs["选择颜色"] = c

                        spec_key = json.dumps(specs, ensure_ascii=False, sort_keys=True)
                        h = hashlib.md5(spec_key.encode("utf-8")).hexdigest()[:8]
                        sku_code = f"{spu_code}-{h}"
                        sku_name = f"{pname} {v} {c}".strip()

                        self.upsert_sku(
                            product_id=pid,
                            sku_code=sku_code,
                            name=sku_name,
                            specs=specs,
                            price=base_price,
                            stock=stock,
                            image=image,
                            status=1,
                        )
                        total += 1

                print(f"  ✓ 写入/更新 SKU 数量: {total}（版本={len(versions)} 颜色={len(colors)}）")
            except Exception as e:
                print(f"  ⚠️  爬取 SKU 失败: {e}")
                import traceback
                traceback.print_exc()
                continue
    
    def download_image(self, img_url: str, save_dir: Path) -> Optional[str]:
        """下载图片"""
        try:
            # 确保URL完整
            if not img_url.startswith('http'):
                img_url = urljoin('https://www.mi.com', img_url)
            
            # 移除URL参数中的尺寸限制（参考Go版本）
            # 小米商城的图片URL可能包含 ?thumb=1&w=40&h=40&f=webp&q=90 等参数
            if '?' in img_url or '&' in img_url:
                # 移除所有尺寸和压缩参数
                img_url = re.sub(r'[?&](thumb|w|h|f|q)=\d+', '', img_url)
                img_url = img_url.rstrip('?&')
                # 如果还有?，说明有其他参数，直接移除所有参数
                if '?' in img_url:
                    img_url = img_url.split('?')[0]
            
            # 下载图片（增加重试，解决 Connection reset by peer）
            last_err = None
            response = None
            for attempt in range(1, 4):
                try:
                    response = self.session.get(img_url, timeout=(10, 30))
                    response.raise_for_status()
                    last_err = None
                    break
                except Exception as e:
                    last_err = e
                    wait_s = 0.8 * attempt
                    print(f"        下载失败(第{attempt}次): {e}，{wait_s:.1f}s 后重试...")
                    time.sleep(wait_s)
                    continue
            if last_err is not None or response is None:
                raise last_err or Exception("download failed")
            
            # 检查Content-Type
            content_type = response.headers.get('Content-Type', '')
            if not content_type.startswith('image/'):
                print(f"        警告: 不是图片类型: {content_type}")
                return None
            
            # 确定文件扩展名
            ext = Path(img_url).suffix
            if not ext:
                if 'png' in content_type:
                    ext = '.png'
                elif 'jpeg' in content_type or 'jpg' in content_type:
                    ext = '.jpg'
                elif 'webp' in content_type:
                    ext = '.webp'
                else:
                    ext = '.jpg'
            
            # 生成文件名
            url_hash = hashlib.md5(img_url.encode()).hexdigest()[:16]
            filename = f"{url_hash}_{int(time.time() * 1000000)}{ext}"
            filepath = save_dir / filename
            
            # 保存文件
            with open(filepath, 'wb') as f:
                f.write(response.content)
            
            file_size = filepath.stat().st_size
            print(f"        文件大小: {file_size / 1024:.2f} KB")
            
            # 返回相对于 ecommerce-backend 目录的路径
            # 查找 ecommerce-backend 目录
            current = Path.cwd()
            backend_dir = None
            for parent in [current] + list(current.parents):
                if parent.name == 'ecommerce-backend':
                    backend_dir = parent
                    break
            
            if backend_dir:
                try:
                    rel_path = filepath.relative_to(backend_dir)
                    # 确保路径使用正斜杠（跨平台兼容）
                    return str(rel_path).replace('\\', '/')
                except ValueError:
                    # 如果不在 backend_dir 下，返回绝对路径
                    return str(filepath).replace('\\', '/')
            else:
                # 如果找不到 backend_dir，返回相对于 image_dir 的路径
                try:
                    rel_path = filepath.relative_to(self.image_dir)
                    return f"images/mi/{rel_path}".replace('\\', '/')
                except ValueError:
                    # 最后返回绝对路径
                    return str(filepath).replace('\\', '/')
        
        except Exception as e:
            print(f"        下载图片失败: {e}")
            return None
    
    def close(self):
        """关闭资源"""
        if self.driver:
            self.driver.quit()
        if self.db:
            self.db.close()
        print("资源已释放")


def main():
    parser = argparse.ArgumentParser(description='小米商城爬虫')
    parser.add_argument('--db-host', default='localhost', help='数据库主机')
    parser.add_argument('--db-port', type=int, default=3306, help='数据库端口')
    parser.add_argument('--db-user', default='root', help='数据库用户')
    parser.add_argument('--db-password', default='123456', help='数据库密码')
    parser.add_argument('--db-name', default='ecommerce', help='数据库名称')
    parser.add_argument('--image-dir', default='images/mi', help='图片保存目录')
    parser.add_argument('--redis-host', default='localhost', help='Redis 主机（用于清理分类树缓存）')
    parser.add_argument('--redis-port', type=int, default=6379, help='Redis 端口')
    parser.add_argument('--redis-password', default='', help='Redis 密码')
    parser.add_argument('--redis-db', type=int, default=0, help='Redis DB')
    parser.add_argument('--only-main-categories', action='store_true', help='只爬主分类（分类阶段）')
    parser.add_argument('--stages', default='categories,products,skus', help='执行阶段：categories,products,skus（逗号分隔）')
    parser.add_argument('--sku-limit', type=int, default=0, help='SKU阶段：限制处理的商品数量（0=不限制）')
    parser.add_argument('--max-subcategories-per-parent', type=int, default=5, help='每个主分类最多处理的子分类数量（默认5）')
    parser.add_argument('--max-products-per-subcategory', type=int, default=50, help='每个子分类最多处理的商品数量（默认50）')
    
    args = parser.parse_args()
    
    db_config = {
        'host': args.db_host,
        'port': args.db_port,
        'user': args.db_user,
        'password': args.db_password,
        'database': args.db_name,
        'redis_host': args.redis_host,
        'redis_port': args.redis_port,
        'redis_password': args.redis_password,
        'redis_db': args.redis_db,
    }
    
    crawler = MiCrawler(db_config, args.image_dir)
    
    try:
        crawler.max_subcategories_per_parent = max(1, int(args.max_subcategories_per_parent))
        crawler.max_products_per_subcategory = max(1, int(args.max_products_per_subcategory))

        # 每次运行强制清库+删图（按你的要求：每次重新爬都清空之前数据）
        crawler.reset_database_and_images(truncate_tables=True, delete_images=True)

        stages = [s.strip() for s in (args.stages or "").split(",") if s.strip()]
        if "categories" in stages:
            crawler.crawl_categories(only_main_categories=args.only_main_categories)

        # 现在的 crawl_categories 在旧 DOM 模式下仍会顺带爬商品；后续你确认新 DOM 子分类结构后，
        # 我们再把 products 阶段完全拆出来（按分类ID+URL爬）
        if "skus" in stages:
            crawler.crawl_skus_for_products(limit=args.sku_limit)
    finally:
        crawler.close()


if __name__ == '__main__':
    main()

