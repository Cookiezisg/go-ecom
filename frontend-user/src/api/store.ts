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
  Sku,
  SkuListData,
  SeckillActivity,
  SeckillActivityListData,
  TrackingNode,
  UserCoupon,
  UserProfile,
  LogisticsInfo,
} from "@/types";

function pickNumber(value: unknown, fallback = 0) {
  if (typeof value === "number") return value;
  if (typeof value === "string" && value !== "") return Number(value);
  return fallback;
}

function pickString(value: unknown, fallback = "") {
  return typeof value === "string" ? value : fallback;
}

function trimTrailingSlash(value: string) {
  return value.endsWith("/") ? value.slice(0, -1) : value;
}

function getGatewayOrigin() {
  const assetBase = pickString(import.meta.env.VITE_ASSET_BASE_URL);
  if (assetBase) {
    return trimTrailingSlash(assetBase);
  }

  const apiBase = pickString(import.meta.env.VITE_API_BASE_URL);
  if (apiBase) {
    if (apiBase.startsWith("http://") || apiBase.startsWith("https://")) {
      return trimTrailingSlash(apiBase);
    }
    return "";
  }

  if (typeof window !== "undefined") {
    const { protocol, hostname, port, origin } = window.location;
    if (port === "5173" || port === "5174") {
      return `${protocol}//${hostname}:8080`;
    }
    return origin;
  }

  return "";
}

export function resolveAssetUrl(value?: string) {
  if (!value) return "";
  if (value.startsWith("http://") || value.startsWith("https://")) return value;
  const normalized = value.startsWith("/") ? value : `/${value}`;
  const gatewayOrigin = getGatewayOrigin();
  return gatewayOrigin ? `${gatewayOrigin}${normalized}` : normalized;
}

function normalizeProduct(input: Record<string, unknown>): Product {
  return {
    id: pickNumber(input.id),
    name: pickString(input.name),
    subtitle: pickString(input.subtitle),
    main_image: resolveAssetUrl(pickString(input.main_image ?? input.mainImage)),
    local_main_image: resolveAssetUrl(pickString(input.local_main_image ?? input.localMainImage)),
    price: pickNumber(input.price),
    original_price: pickNumber(input.original_price ?? input.originalPrice),
    stock: pickNumber(input.stock),
    sales: pickNumber(input.sales),
    detail: pickString(input.detail),
    is_hot: pickNumber(input.is_hot ?? input.isHot),
  };
}

function normalizeSearchProduct(input: Record<string, unknown>): SearchProductResult {
  return {
    product_id: pickNumber(input.product_id ?? input.productId),
    name: pickString(input.name),
    main_image: resolveAssetUrl(pickString(input.main_image ?? input.mainImage)),
    price: pickString(input.price, String(pickNumber(input.price))),
    sales: pickNumber(input.sales),
    score: pickNumber(input.score),
  };
}

function normalizeCartItem(input: Record<string, unknown>): CartItem {
  return {
    id: pickNumber(input.id),
    user_id: pickNumber(input.user_id ?? input.userId),
    sku_id: pickNumber(input.sku_id ?? input.skuId),
    quantity: pickNumber(input.quantity, 1),
    product_id: pickNumber(input.product_id ?? input.productId),
    product_name: pickString(input.product_name ?? input.productName),
    sku_name: pickString(input.sku_name ?? input.skuName),
    sku_image: resolveAssetUrl(pickString(input.sku_image ?? input.skuImage)),
    price: pickString(input.price, "0"),
    stock_status: pickString(input.stock_status ?? input.stockStatus),
    is_selected: pickNumber(input.is_selected ?? input.isSelected),
  };
}

function normalizeSku(input: Record<string, unknown>): Sku {
  const rawSpecs = input.specs;
  const specs =
    rawSpecs && typeof rawSpecs === "object" && !Array.isArray(rawSpecs)
      ? Object.fromEntries(
          Object.entries(rawSpecs as Record<string, unknown>).map(([key, value]) => [key, pickString(value)]),
        )
      : {};
  return {
    id: pickNumber(input.id),
    product_id: pickNumber(input.product_id ?? input.productId),
    sku_code: pickString(input.sku_code ?? input.skuCode),
    name: pickString(input.name),
    specs,
    price: pickNumber(input.price),
    original_price: pickNumber(input.original_price ?? input.originalPrice),
    stock: pickNumber(input.stock),
    image: resolveAssetUrl(pickString(input.image)),
    status: pickNumber(input.status),
  };
}

function normalizeUserProfile(input: Record<string, unknown>): UserProfile {
  return {
    id: pickNumber(input.id),
    username: pickString(input.username),
    nickname: pickString(input.nickname),
    phone: pickString(input.phone),
    email: pickString(input.email),
    avatar: resolveAssetUrl(pickString(input.avatar)),
    gender: pickNumber(input.gender),
    birthday: pickString(input.birthday),
    status: pickNumber(input.status),
    member_level: pickNumber(input.member_level ?? input.memberLevel),
    points: pickNumber(input.points),
  };
}

function normalizeAddress(input: Record<string, unknown>): Address {
  return {
    id: pickNumber(input.id),
    user_id: pickNumber(input.user_id ?? input.userId),
    receiver_name: pickString(input.receiver_name ?? input.receiverName),
    receiver_phone: pickString(input.receiver_phone ?? input.receiverPhone),
    province: pickString(input.province),
    city: pickString(input.city),
    district: pickString(input.district),
    detail: pickString(input.detail),
    postal_code: pickString(input.postal_code ?? input.postalCode),
    is_default: pickNumber(input.is_default ?? input.isDefault),
  };
}

function normalizeReview(input: Record<string, unknown>): Review {
  return {
    id: pickNumber(input.id),
    user_id: pickNumber(input.user_id ?? input.userId),
    content: pickString(input.content),
    rating: pickNumber(input.rating),
    created_at: pickString(input.created_at ?? input.createdAt),
    reply_content: pickString(input.reply_content ?? input.replyContent),
  };
}

function normalizeReviewStats(input: Record<string, unknown>): ReviewStats {
  return {
    total_count: pickNumber(input.total_count ?? input.totalCount),
    average_rating: typeof input.average_rating === "number"
      ? input.average_rating
      : typeof input.averageRating === "number"
        ? input.averageRating
        : pickNumber(input.average_rating ?? input.averageRating),
  };
}

function normalizeOrderItem(input: Record<string, unknown>): Order["items"][number] {
  return {
    id: pickNumber(input.id),
    order_id: pickNumber(input.order_id ?? input.orderId),
    order_no: pickString(input.order_no ?? input.orderNo),
    product_id: pickNumber(input.product_id ?? input.productId),
    product_name: pickString(input.product_name ?? input.productName),
    sku_id: pickNumber(input.sku_id ?? input.skuId),
    sku_name: pickString(input.sku_name ?? input.skuName),
    sku_image: resolveAssetUrl(pickString(input.sku_image ?? input.skuImage)),
    price: pickString(input.price, "0"),
    quantity: pickNumber(input.quantity),
    total_amount: pickString(input.total_amount ?? input.totalAmount, "0"),
  };
}

function normalizeOrder(input: Record<string, unknown>): Order {
  return {
    id: pickNumber(input.id),
    order_no: pickString(input.order_no ?? input.orderNo),
    user_id: pickNumber(input.user_id ?? input.userId),
    status: pickNumber(input.status),
    total_amount: pickString(input.total_amount ?? input.totalAmount, "0"),
    pay_amount: pickString(input.pay_amount ?? input.payAmount, "0"),
    discount_amount: pickString(input.discount_amount ?? input.discountAmount),
    freight_amount: pickString(input.freight_amount ?? input.freightAmount),
    receiver_name: pickString(input.receiver_name ?? input.receiverName),
    receiver_phone: pickString(input.receiver_phone ?? input.receiverPhone),
    receiver_address: pickString(input.receiver_address ?? input.receiverAddress),
    remark: pickString(input.remark),
    items: (Array.isArray(input.items) ? input.items : []).map((item) =>
      normalizeOrderItem(item as Record<string, unknown>),
    ),
    created_at: pickString(input.created_at ?? input.createdAt),
    updated_at: pickString(input.updated_at ?? input.updatedAt),
  };
}

function normalizeCoupon(input: Record<string, unknown>): Coupon {
  return {
    id: pickNumber(input.id),
    name: pickString(input.name),
    type: pickNumber(input.type),
    discount_type: pickNumber(input.discount_type ?? input.discountType),
    discount_value: pickString(input.discount_value ?? input.discountValue, "0"),
    min_amount: pickString(input.min_amount ?? input.minAmount),
    max_discount: pickString(input.max_discount ?? input.maxDiscount),
    valid_start_time: pickString(input.valid_start_time ?? input.validStartTime),
    valid_end_time: pickString(input.valid_end_time ?? input.validEndTime),
    status: pickNumber(input.status),
  };
}

function normalizeUserCoupon(input: Record<string, unknown>): UserCoupon {
  return {
    id: pickNumber(input.id),
    user_id: pickNumber(input.user_id ?? input.userId),
    coupon_id: pickNumber(input.coupon_id ?? input.couponId),
    status: pickNumber(input.status),
    order_id: pickNumber(input.order_id ?? input.orderId),
    expire_at: pickString(input.expire_at ?? input.expireAt),
    created_at: pickString(input.created_at ?? input.createdAt),
  };
}

function normalizeSeckillActivity(input: Record<string, unknown>): SeckillActivity {
  return {
    id: pickNumber(input.id),
    name: pickString(input.name),
    sku_id: pickNumber(input.sku_id ?? input.skuId),
    sku_name: pickString(input.sku_name ?? input.skuName),
    sku_image: resolveAssetUrl(pickString(input.sku_image ?? input.skuImage)),
    seckill_price: pickString(input.seckill_price ?? input.seckillPrice, "0"),
    original_price: pickString(input.original_price ?? input.originalPrice, "0"),
    stock: pickNumber(input.stock),
    sold: pickNumber(input.sold),
    start_time: pickNumber(input.start_time ?? input.startTime),
    end_time: pickNumber(input.end_time ?? input.endTime),
    status: pickNumber(input.status),
    enable_status: pickNumber(input.enable_status ?? input.enableStatus),
  };
}

function normalizeLogisticsInfo(input: Record<string, unknown>): LogisticsInfo {
  return {
    id: pickNumber(input.id),
    order_id: pickNumber(input.order_id ?? input.orderId),
    order_no: pickString(input.order_no ?? input.orderNo),
    logistics_no: pickString(input.logistics_no ?? input.logisticsNo),
    company_code: pickString(input.company_code ?? input.companyCode),
    company_name: pickString(input.company_name ?? input.companyName),
    status: pickNumber(input.status),
    receiver_name: pickString(input.receiver_name ?? input.receiverName),
    receiver_phone: pickString(input.receiver_phone ?? input.receiverPhone),
    receiver_address: pickString(input.receiver_address ?? input.receiverAddress),
  };
}

function normalizeTrackingNode(input: Record<string, unknown>): TrackingNode {
  return {
    time: pickString(input.time),
    status: pickString(input.status),
    location: pickString(input.location),
    remark: pickString(input.remark),
  };
}

export interface LoginFormValues {
  username: string;
  password: string;
  login_type: number;
}

export async function login(values: LoginFormValues) {
  const response = await apiClient.post<ApiResponse<LoginPayload>>("/api/v1/user/login", values);
  const payload = response.data;
  return {
    ...payload,
    data: {
      ...payload.data,
      user: normalizeUserProfile(payload.data.user as unknown as Record<string, unknown>),
      expire_time: pickNumber((payload.data as unknown as Record<string, unknown>).expire_time ?? (payload.data as unknown as Record<string, unknown>).expireTime),
    },
  };
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
  const payload = response.data;
  return {
    ...payload,
    data: {
      ...payload.data,
      list: (payload.data?.list || []).map((item) => normalizeProduct(item as unknown as Record<string, unknown>)),
    },
  };
}

export async function listBanners() {
  const response = await apiClient.get<ApiResponse<Banner[]>>("/api/v1/banners", {
    params: { status: 1, limit: 5 },
  });
  const payload = response.data;
  return {
    ...payload,
    data: (payload.data || []).map((item) => ({
      ...item,
      image: resolveAssetUrl(item.image),
      image_local: resolveAssetUrl(item.image_local),
    })),
  };
}

export async function getProduct(id: string) {
  const response = await apiClient.get<ApiResponse<Product>>(`/api/v1/products/${id}`);
  const payload = response.data;
  return {
    ...payload,
    data: normalizeProduct(payload.data as unknown as Record<string, unknown>),
  };
}

export async function listSkus(params: Record<string, string | number | undefined>) {
  const response = await apiClient.get<ApiResponse<SkuListData>>("/api/v1/skus", { params });
  const payload = response.data;
  const raw = payload.data as unknown as Record<string, unknown>;
  const list = (raw?.list as unknown[] | undefined) || [];
  return {
    ...payload,
    data: {
      list: list.map((item) => normalizeSku(item as Record<string, unknown>)),
      page: pickNumber(raw?.page),
      page_size: pickNumber(raw?.page_size ?? raw?.pageSize),
      total: pickNumber(raw?.total),
      total_pages: pickNumber(raw?.total_pages ?? raw?.totalPages),
    },
  };
}

export async function getProductReviews(id: string) {
  const response = await apiClient.get<ApiResponse<Review[]>>(`/api/v1/reviews/product/${id}`, {
    params: { page: 1, page_size: 10 },
  });
  const payload = response.data;
  return {
    ...payload,
    data: (payload.data || []).map((item) => normalizeReview(item as unknown as Record<string, unknown>)),
  };
}

export async function getReviewStats(id: string) {
  const response = await apiClient.get<ApiResponse<ReviewStats>>(`/api/v1/reviews/stats/${id}`);
  const payload = response.data;
  return {
    ...payload,
    data: normalizeReviewStats(payload.data as unknown as Record<string, unknown>),
  };
}

export async function getCart(userId: number) {
  const response = await apiClient.get<ApiResponse<CartItem[]>>("/api/v1/cart", {
    params: { user_id: userId },
  });
  const payload = response.data;
  return {
    ...payload,
    data: (payload.data || []).map((item) => normalizeCartItem(item as unknown as Record<string, unknown>)),
  };
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
  const payload = response.data;
  return {
    ...payload,
    data: (payload.data || []).map((item) => normalizeAddress(item as unknown as Record<string, unknown>)),
  };
}

export async function addAddress(payload: Omit<Address, "id">) {
  const response = await apiClient.post<ApiResponse<Address>>("/api/v1/user/address", payload);
  const result = response.data;
  return {
    ...result,
    data: normalizeAddress(result.data as unknown as Record<string, unknown>),
  };
}

export async function updateAddress(payload: Address) {
  const response = await apiClient.put<ApiResponse<Address>>(`/api/v1/user/address/${payload.id}`, payload);
  const result = response.data;
  return {
    ...result,
    data: normalizeAddress(result.data as unknown as Record<string, unknown>),
  };
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
  const payload = response.data;
  return {
    ...payload,
    data: normalizeUserProfile(payload.data as unknown as Record<string, unknown>),
  };
}

export async function updateUserInfo(payload: Partial<UserProfile> & { user_id: number }) {
  const response = await apiClient.put<ApiResponse<UserProfile>>("/api/v1/user/info", payload);
  const result = response.data;
  return {
    ...result,
    data: normalizeUserProfile(result.data as unknown as Record<string, unknown>),
  };
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
  const result = response.data;
  return {
    ...result,
    data: normalizeOrder(result.data as unknown as Record<string, unknown>),
  };
}

export async function listOrders(userId: number, status = -1, page = 1) {
  const response = await apiClient.get<ApiResponse<OrderListData>>("/api/v1/orders", {
    params: { user_id: userId, status, page, page_size: 20 },
  });
  const payload = response.data;
  const raw = payload.data as unknown as Record<string, unknown>;
  const list = (raw?.list as unknown[] | undefined) || [];
  return {
    ...payload,
    data: {
      list: list.map((item) => normalizeOrder(item as Record<string, unknown>)),
      page: pickNumber(raw?.page),
      page_size: pickNumber(raw?.page_size ?? raw?.pageSize),
      total: pickNumber(raw?.total),
      total_pages: pickNumber(raw?.total_pages ?? raw?.totalPages),
    },
  };
}

export async function getOrder(id: string) {
  const response = await apiClient.get<ApiResponse<Order>>(`/api/v1/orders/${id}`);
  const payload = response.data;
  return {
    ...payload,
    data: normalizeOrder(payload.data as unknown as Record<string, unknown>),
  };
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
  const result = response.data as unknown as Record<string, unknown>;
  return {
    ...(result as object),
    data: {
      payment_no: pickString((result.data as Record<string, unknown> | undefined)?.payment_no ?? (result.data as Record<string, unknown> | undefined)?.paymentNo),
    },
    pay_url: pickString(result.pay_url ?? result.payUrl),
  };
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
  const payload = response.data;
  return {
    ...payload,
    data: (payload.data || []).map((item) => normalizeCoupon(item as unknown as Record<string, unknown>)),
  };
}

export async function getUserCoupons(userId: number, status = 0) {
  const response = await apiClient.get<ApiResponse<UserCoupon[]>>(`/api/v1/promotion/user-coupons/${userId}`, {
    params: { status },
  });
  const payload = response.data;
  return {
    ...payload,
    data: (payload.data || []).map((item) => normalizeUserCoupon(item as unknown as Record<string, unknown>)),
  };
}

export async function listSeckillActivities(status = 1) {
  const response = await apiClient.get<ApiResponse<SeckillActivityListData>>("/api/v1/seckill/activities", {
    params: { page: 1, page_size: 20, status },
  });
  const payload = response.data;
  return {
    ...payload,
    data: {
      ...payload.data,
      list: (payload.data?.list || []).map((item) => normalizeSeckillActivity(item as unknown as Record<string, unknown>)),
    },
  };
}

export async function getSeckillActivity(id: string) {
  const response = await apiClient.get<ApiResponse<SeckillActivity>>(`/api/v1/seckill/activities/${id}`);
  const payload = response.data;
  return {
    ...payload,
    data: normalizeSeckillActivity(payload.data as unknown as Record<string, unknown>),
  };
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
  const result = response.data;
  return {
    ...result,
    data: normalizeReview(result.data as unknown as Record<string, unknown>),
  };
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
  const payload = response.data;
  return {
    ...payload,
    data: (payload.data || []).map((item) => normalizeSearchProduct(item as unknown as Record<string, unknown>)),
  };
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
  const payload = response.data;
  return {
    ...payload,
    data: normalizeLogisticsInfo(payload.data as unknown as Record<string, unknown>),
  };
}

export async function getTracking(logisticsNo: string) {
  const response = await apiClient.get<ApiResponse<TrackingNode[]>>(`/api/v1/logistics/tracking/${logisticsNo}`);
  const payload = response.data;
  return {
    ...payload,
    data: (payload.data || []).map((item) => normalizeTrackingNode(item as unknown as Record<string, unknown>)),
  };
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
  const payload = response.data as unknown as Record<string, unknown>;
  const data = (payload.data as Record<string, unknown> | undefined) || {};
  return {
    ...(payload as object),
    data: {
      file_url: resolveAssetUrl(pickString(data.file_url ?? data.fileUrl)),
      file_name: pickString(data.file_name ?? data.fileName),
      file_id: pickString(data.file_id ?? data.fileId),
    },
  };
}
