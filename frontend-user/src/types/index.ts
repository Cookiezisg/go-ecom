export interface ApiResponse<T> {
  code: number;
  message: string;
  data: T;
}

export interface UserProfile {
  id: number;
  username: string;
  nickname: string;
  phone?: string;
  email?: string;
  avatar?: string;
  gender?: number;
  birthday?: string;
  status?: number;
  member_level?: number;
  points?: number;
}

export interface LoginPayload {
  user: UserProfile;
  token: string;
  expire_time: number;
}

export interface Product {
  id: number;
  name: string;
  subtitle?: string;
  main_image?: string;
  local_main_image?: string;
  price: number;
  original_price?: number;
  stock: number;
  sales: number;
  detail?: string;
  is_hot?: number;
}

export interface ProductListData {
  list: Product[];
  page: number;
  page_size: number;
  total: number;
  total_pages: number;
}

export interface Review {
  id: number;
  user_id: number;
  content: string;
  rating: number;
  created_at: string;
  reply_content?: string;
}

export interface ReviewStats {
  total_count: number;
  average_rating: number;
}

export interface CartItem {
  id: number;
  user_id: number;
  sku_id: number;
  quantity: number;
  product_id: number;
  product_name: string;
  sku_name: string;
  sku_image?: string;
  price: string;
  stock_status: string;
  is_selected: number;
}

export interface Address {
  id: number;
  user_id: number;
  receiver_name: string;
  receiver_phone: string;
  province: string;
  city: string;
  district: string;
  detail: string;
  postal_code?: string;
  is_default: number;
}

export interface OrderItem {
  id: number;
  order_id: number;
  order_no: string;
  product_id: number;
  product_name: string;
  sku_id: number;
  sku_name: string;
  sku_image?: string;
  price: string;
  quantity: number;
  total_amount: string;
}

export interface Order {
  id: number;
  order_no: string;
  user_id: number;
  status: number;
  total_amount: string;
  pay_amount: string;
  discount_amount?: string;
  freight_amount?: string;
  receiver_name?: string;
  receiver_phone?: string;
  receiver_address?: string;
  remark?: string;
  items: OrderItem[];
  created_at?: string;
  updated_at?: string;
}

export interface OrderListData {
  list: Order[];
  page: number;
  page_size: number;
  total: number;
  total_pages: number;
}

export interface Coupon {
  id: number;
  name: string;
  type: number;
  discount_type: number;
  discount_value: string;
  min_amount?: string;
  max_discount?: string;
  valid_start_time?: string;
  valid_end_time?: string;
  status: number;
}

export interface UserCoupon {
  id: number;
  user_id: number;
  coupon_id: number;
  status: number;
  order_id?: number;
  expire_at?: string;
  created_at?: string;
}

export interface SeckillActivity {
  id: number;
  name: string;
  sku_id: number;
  sku_name: string;
  sku_image?: string;
  seckill_price: string;
  original_price: string;
  stock: number;
  sold: number;
  start_time: number;
  end_time: number;
  status: number;
  enable_status: number;
}

export interface SeckillActivityListData {
  list: SeckillActivity[];
  page: number;
  page_size: number;
  total: number;
  total_pages: number;
}

export interface Banner {
  id: number;
  title: string;
  description?: string;
  image?: string;
  image_local?: string;
  link?: string;
  link_type?: number;
}

export interface SearchProductResult {
  product_id: number;
  name: string;
  main_image?: string;
  price: string;
  sales: number;
  score: number;
}

export interface LogisticsInfo {
  id: number;
  order_id: number;
  order_no: string;
  logistics_no: string;
  company_code?: string;
  company_name?: string;
  status: number;
  receiver_name?: string;
  receiver_phone?: string;
  receiver_address?: string;
}

export interface TrackingNode {
  time: string;
  status: string;
  location: string;
  remark: string;
}
