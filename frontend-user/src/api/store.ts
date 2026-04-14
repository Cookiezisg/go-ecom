import { apiClient } from "@/api/client";
import type {
  Address,
  ApiResponse,
  Banner,
  CartItem,
  Coupon,
  LoginPayload,
  Order,
  OrderListData,
  Product,
  ProductListData,
  Review,
  ReviewStats,
  SearchProductResult,
  SeckillActivity,
  SeckillActivityListData,
  TrackingNode,
  UserCoupon,
  UserProfile,
  LogisticsInfo,
} from "@/types";

export interface LoginFormValues {
  username: string;
  password: string;
  login_type: number;
}

export async function login(values: LoginFormValues) {
  const response = await apiClient.post<ApiResponse<LoginPayload>>("/api/v1/user/login", values);
  return response.data;
}

export async function register(values: {
  username: string;
  password: string;
  phone: string;
  email: string;
}) {
  const response = await apiClient.post<ApiResponse<unknown>>("/api/v1/user/register", values);
  return response.data;
}

export async function listProducts(params: Record<string, string | number | undefined>) {
  const response = await apiClient.get<ApiResponse<ProductListData>>("/api/v1/products", { params });
  return response.data;
}

export async function listBanners() {
  const response = await apiClient.get<ApiResponse<Banner[]>>("/api/v1/banners", {
    params: { status: 1, limit: 5 },
  });
  return response.data;
}

export async function getProduct(id: string) {
  const response = await apiClient.get<ApiResponse<Product>>(`/api/v1/products/${id}`);
  return response.data;
}

export async function getProductReviews(id: string) {
  const response = await apiClient.get<ApiResponse<Review[]>>(`/api/v1/reviews/product/${id}`, {
    params: { page: 1, page_size: 10 },
  });
  return response.data;
}

export async function getReviewStats(id: string) {
  const response = await apiClient.get<ApiResponse<ReviewStats>>(`/api/v1/reviews/stats/${id}`);
  return response.data;
}

export async function getCart(userId: number) {
  const response = await apiClient.get<ApiResponse<CartItem[]>>("/api/v1/cart", {
    params: { user_id: userId },
  });
  return response.data;
}

export async function addCartItem(payload: { user_id: number; sku_id: number; quantity: number }) {
  const response = await apiClient.post<ApiResponse<CartItem>>("/api/v1/cart", payload);
  return response.data;
}

export async function updateCartQuantity(payload: {
  user_id: number;
  sku_id: number;
  quantity: number;
}) {
  const response = await apiClient.put<ApiResponse<unknown>>(`/api/v1/cart/${payload.sku_id}`, payload);
  return response.data;
}

export async function removeCartItems(userId: number, skuIds: number[]) {
  const response = await apiClient.delete<ApiResponse<unknown>>(`/api/v1/cart/${skuIds.join(",")}`, {
    data: { user_id: userId, sku_ids: skuIds },
  });
  return response.data;
}

export async function getAddresses(userId: number) {
  const response = await apiClient.get<ApiResponse<Address[]>>("/api/v1/user/address", {
    params: { user_id: userId },
  });
  return response.data;
}

export async function addAddress(payload: Omit<Address, "id">) {
  const response = await apiClient.post<ApiResponse<Address>>("/api/v1/user/address", payload);
  return response.data;
}

export async function updateAddress(payload: Address) {
  const response = await apiClient.put<ApiResponse<Address>>(`/api/v1/user/address/${payload.id}`, payload);
  return response.data;
}

export async function deleteAddress(id: number, userId: number) {
  const response = await apiClient.delete<ApiResponse<unknown>>(`/api/v1/user/address/${id}`, {
    data: { id, user_id: userId },
  });
  return response.data;
}

export async function getUserInfo(userId: number) {
  const response = await apiClient.get<ApiResponse<UserProfile>>("/api/v1/user/info", {
    params: { user_id: userId },
  });
  return response.data;
}

export async function updateUserInfo(payload: Partial<UserProfile> & { user_id: number }) {
  const response = await apiClient.put<ApiResponse<UserProfile>>("/api/v1/user/info", payload);
  return response.data;
}

export async function createOrder(payload: {
  user_id: number;
  address_id: number;
  items: Array<{ sku_id: number; quantity: number }>;
  order_type: number;
  coupon_id?: number;
  remark?: string;
}) {
  const response = await apiClient.post<ApiResponse<Order>>("/api/v1/orders", payload);
  return response.data;
}

export async function listOrders(userId: number, status = -1, page = 1) {
  const response = await apiClient.get<ApiResponse<OrderListData>>("/api/v1/orders", {
    params: { user_id: userId, status, page, page_size: 20 },
  });
  return response.data;
}

export async function getOrder(id: string) {
  const response = await apiClient.get<ApiResponse<Order>>(`/api/v1/orders/${id}`);
  return response.data;
}

export async function cancelOrder(id: number, reason: string) {
  const response = await apiClient.put<ApiResponse<unknown>>(`/api/v1/orders/${id}/cancel`, { id, reason });
  return response.data;
}

export async function createPayment(payload: {
  order_id: number;
  order_no: string;
  user_id: number;
  amount: string;
  payment_method: number;
}) {
  const response = await apiClient.post<ApiResponse<{ payment_no: string }> & { pay_url?: string }>(
    "/api/v1/payments",
    payload,
  );
  return response.data;
}

export async function getPayment(paymentNo: string) {
  const response = await apiClient.get<ApiResponse<Record<string, unknown>>>(`/api/v1/payments/${paymentNo}`);
  return response.data;
}

export async function getPaymentStatus(paymentNo: string) {
  const response = await apiClient.get<ApiResponse<number>>(`/api/v1/payments/${paymentNo}/status`);
  return response.data;
}

export async function listCoupons() {
  const response = await apiClient.get<ApiResponse<Coupon[]>>("/api/v1/promotion/coupons", {
    params: { page: 1, page_size: 20 },
  });
  return response.data;
}

export async function getUserCoupons(userId: number, status = 0) {
  const response = await apiClient.get<ApiResponse<UserCoupon[]>>(`/api/v1/promotion/user-coupons/${userId}`, {
    params: { status },
  });
  return response.data;
}

export async function listSeckillActivities(status = 1) {
  const response = await apiClient.get<ApiResponse<SeckillActivityListData>>("/api/v1/seckill/activities", {
    params: { page: 1, page_size: 20, status },
  });
  return response.data;
}

export async function getSeckillActivity(id: string) {
  const response = await apiClient.get<ApiResponse<SeckillActivity>>(`/api/v1/seckill/activities/${id}`);
  return response.data;
}

export async function seckill(payload: { user_id: number; sku_id: number; quantity: number }) {
  const response = await apiClient.post<ApiResponse<{ success: boolean; order_no?: string; message?: string }>>(
    "/api/v1/seckill",
    payload,
  );
  return response.data;
}

export async function createReview(payload: {
  user_id: number;
  order_id: number;
  order_item_id: number;
  product_id: number;
  sku_id: number;
  rating: number;
  content: string;
  images?: string[];
  videos?: string[];
}) {
  const response = await apiClient.post<ApiResponse<Review>>("/api/v1/reviews", payload);
  return response.data;
}

export async function searchProducts(params: {
  keyword: string;
  page?: number;
  page_size?: number;
  category_id?: number;
  sort_by?: string;
}) {
  const response = await apiClient.get<ApiResponse<SearchProductResult[]> & { total?: number }>(
    "/api/v1/search/products",
    { params },
  );
  return response.data;
}

export async function getSearchSuggestions(keyword: string, limit = 8) {
  const response = await apiClient.get<ApiResponse<string[]>>("/api/v1/search/suggestions", {
    params: { keyword, limit },
  });
  return response.data;
}

export async function getHotKeywords(limit = 10) {
  const response = await apiClient.get<ApiResponse<string[]>>("/api/v1/search/hot-keywords", {
    params: { limit },
  });
  return response.data;
}

export async function getLogistics(orderId: number) {
  const response = await apiClient.get<ApiResponse<LogisticsInfo>>(`/api/v1/logistics/${orderId}`);
  return response.data;
}

export async function getTracking(logisticsNo: string) {
  const response = await apiClient.get<ApiResponse<TrackingNode[]>>(`/api/v1/logistics/tracking/${logisticsNo}`);
  return response.data;
}

export async function confirmReceive(orderId: number, orderNo?: string) {
  const response = await apiClient.put<ApiResponse<unknown>>(`/api/v1/orders/${orderId}/confirm-receive`, {
    id: orderId,
    order_no: orderNo,
  });
  return response.data;
}

export async function uploadFile(file: File, category = "image") {
  const formData = new FormData();
  formData.append("file", file);
  formData.append("category", category);
  const response = await apiClient.post<ApiResponse<{ file_url?: string; file_name?: string; file_id?: string }>>(
    "/api/v1/files/upload",
    formData,
    {
      headers: {
        "Content-Type": "multipart/form-data",
      },
      timeout: 60000,
    },
  );
  return response.data;
}
