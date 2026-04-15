import { apiClient } from "@/api/client";

export interface ApiResponse<T> {
  code: number;
  message: string;
  data: T;
}

function pickNumber(value: unknown, fallback = 0) {
  if (typeof value === "number") return value;
  if (typeof value === "string" && value !== "") return Number(value);
  return fallback;
}

function pickString(value: unknown, fallback = "") {
  return typeof value === "string" ? value : fallback;
}

function normalizeProduct(input: Record<string, unknown>) {
  return {
    id: pickNumber(input.id),
    name: pickString(input.name),
    subtitle: pickString(input.subtitle),
    category_id: pickNumber(input.category_id ?? input.categoryId),
    brand_id: pickNumber(input.brand_id ?? input.brandId),
    main_image: pickString(input.main_image ?? input.mainImage),
    local_main_image: pickString(input.local_main_image ?? input.localMainImage),
    detail: pickString(input.detail),
    price: pickNumber(input.price),
    original_price: pickNumber(input.original_price ?? input.originalPrice),
    stock: pickNumber(input.stock),
    sales: pickNumber(input.sales),
    status: pickNumber(input.status),
    is_hot: pickNumber(input.is_hot ?? input.isHot),
  };
}

function normalizeUser(input: Record<string, unknown>) {
  return {
    id: pickNumber(input.id),
    username: pickString(input.username),
    nickname: pickString(input.nickname),
    phone: pickString(input.phone),
    email: pickString(input.email),
    status: pickNumber(input.status),
    member_level: pickNumber(input.member_level ?? input.memberLevel),
    created_at: pickString(input.created_at ?? input.createdAt),
  };
}

function normalizeSku(input: Record<string, unknown>) {
  return {
    id: pickNumber(input.id),
    product_id: pickNumber(input.product_id ?? input.productId),
    sku_code: pickString(input.sku_code ?? input.skuCode),
    name: pickString(input.name),
    price: pickNumber(input.price),
    original_price: pickNumber(input.original_price ?? input.originalPrice),
    stock: pickNumber(input.stock),
    image: pickString(input.image),
    status: pickNumber(input.status),
  };
}

function normalizeCategory(input: Record<string, unknown>): Record<string, unknown> {
  const children = Array.isArray(input.children)
    ? input.children.map((item) => normalizeCategory(item as Record<string, unknown>))
    : [];
  return {
    ...input,
    id: pickNumber(input.id),
    parent_id: pickNumber(input.parent_id ?? input.parentId),
    name: pickString(input.name),
    level: pickNumber(input.level),
    sort: pickNumber(input.sort),
    icon: pickString(input.icon),
    image: pickString(input.image),
    description: pickString(input.description),
    status: pickNumber(input.status),
    children,
  };
}

function normalizeBanner(input: Record<string, unknown>) {
  return {
    id: pickNumber(input.id),
    title: pickString(input.title),
    description: pickString(input.description),
    image: pickString(input.image),
    image_local: pickString(input.image_local ?? input.imageLocal),
    link: pickString(input.link),
    link_type: pickNumber(input.link_type ?? input.linkType),
    sort: pickNumber(input.sort),
    status: pickNumber(input.status),
    start_time: pickString(input.start_time ?? input.startTime),
    end_time: pickString(input.end_time ?? input.endTime),
  };
}

function normalizeOrder(input: Record<string, unknown>) {
  return {
    id: pickNumber(input.id),
    order_no: pickString(input.order_no ?? input.orderNo),
    user_id: pickNumber(input.user_id ?? input.userId),
    status: pickNumber(input.status),
    pay_amount: pickString(input.pay_amount ?? input.payAmount),
    receiver_name: pickString(input.receiver_name ?? input.receiverName),
    created_at: pickString(input.created_at ?? input.createdAt),
  };
}

function normalizeSeckillActivity(input: Record<string, unknown>) {
  return {
    id: pickNumber(input.id),
    name: pickString(input.name),
    sku_id: pickNumber(input.sku_id ?? input.skuId),
    sku_name: pickString(input.sku_name ?? input.skuName),
    seckill_price: pickString(input.seckill_price ?? input.seckillPrice),
    original_price: pickString(input.original_price ?? input.originalPrice),
    stock: pickNumber(input.stock),
    sold: pickNumber(input.sold),
    status: pickNumber(input.status),
    enable_status: pickNumber(input.enable_status ?? input.enableStatus),
    start_time: pickNumber(input.start_time ?? input.startTime),
    end_time: pickNumber(input.end_time ?? input.endTime),
  };
}

export async function adminLogin(values: {
  username: string;
  password: string;
  login_type: number;
}) {
  const response = await apiClient.post<
    ApiResponse<{
      token: string;
      user: { id: number; username: string; nickname?: string; member_level?: number };
    }>
  >("/api/v1/user/login", values);
  const payload = response.data;
  return {
    ...payload,
    data: {
      ...payload.data,
      user: normalizeUser(payload.data.user as unknown as Record<string, unknown>),
    },
  };
}

export async function listUsers(params: Record<string, string | number | undefined>) {
  const response = await apiClient.get<
    ApiResponse<{
      users: Array<{
        id: number;
        username: string;
        nickname?: string;
        phone?: string;
        email?: string;
        status: number;
        member_level?: number;
        created_at?: string;
      }>;
      total: number;
      page: number;
      page_size: number;
    }>
  >("/api/v1/users", { params });
  const payload = response.data;
  return {
    ...payload,
    data: {
      ...payload.data,
      users: (payload.data?.users || []).map((item) => normalizeUser(item as unknown as Record<string, unknown>)),
      page_size: pickNumber((payload.data as Record<string, unknown>)?.page_size ?? (payload.data as Record<string, unknown>)?.pageSize),
    },
  };
}

export async function listProducts(params: Record<string, string | number | undefined>) {
  const response = await apiClient.get<
    ApiResponse<{
      list: Array<{
        id: number;
        name: string;
        subtitle?: string;
        price: number;
        stock: number;
        sales: number;
        status: number;
      }>;
      total: number;
      page: number;
      page_size: number;
      total_pages: number;
    }>
  >("/api/v1/products", { params });
  const payload = response.data;
  return {
    ...payload,
    data: {
      ...payload.data,
      list: (payload.data?.list || []).map((item) => normalizeProduct(item as unknown as Record<string, unknown>)),
    },
  };
}

export async function createProduct(payload: Record<string, unknown>) {
  const response = await apiClient.post<ApiResponse<Record<string, unknown>>>("/api/v1/products", payload);
  return response.data;
}

export async function updateProduct(id: number, payload: Record<string, unknown>) {
  const response = await apiClient.put<ApiResponse<Record<string, unknown>>>(`/api/v1/products/${id}`, payload);
  return response.data;
}

export async function deleteProduct(id: number) {
  const response = await apiClient.delete<ApiResponse<Record<string, unknown>>>(`/api/v1/products/${id}`);
  return response.data;
}

export async function listSkus(params: Record<string, string | number | undefined>) {
  const response = await apiClient.get<
    ApiResponse<{ list: Array<Record<string, unknown>>; total: number; page: number; total_pages: number }>
  >("/api/v1/skus", { params });
  const payload = response.data;
  return {
    ...payload,
    data: {
      ...payload.data,
      list: (payload.data?.list || []).map((item) => normalizeSku(item as Record<string, unknown>)),
      total_pages: pickNumber((payload.data as Record<string, unknown>)?.total_pages ?? (payload.data as Record<string, unknown>)?.totalPages),
    },
  };
}

export async function createSku(payload: Record<string, unknown>) {
  const response = await apiClient.post<ApiResponse<Record<string, unknown>>>("/api/v1/skus", payload);
  return response.data;
}

export async function updateSku(id: number, payload: Record<string, unknown>) {
  const response = await apiClient.put<ApiResponse<Record<string, unknown>>>(`/api/v1/skus/${id}`, payload);
  return response.data;
}

export async function deleteSku(id: number) {
  const response = await apiClient.delete<ApiResponse<Record<string, unknown>>>(`/api/v1/skus/${id}`);
  return response.data;
}

export async function listCategories(params: Record<string, string | number | undefined> = { status: 1 }) {
  const response = await apiClient.get<ApiResponse<Array<Record<string, unknown>>>>("/api/v1/categories", {
    params,
  });
  const payload = response.data;
  return {
    ...payload,
    data: (payload.data || []).map((item) => normalizeCategory(item as Record<string, unknown>)),
  };
}

export async function listCategoryTree(status = 1) {
  const response = await apiClient.get<ApiResponse<Array<Record<string, unknown>>>>("/api/v1/categories/tree", {
    params: { status },
  });
  const payload = response.data;
  return {
    ...payload,
    data: (payload.data || []).map((item) => normalizeCategory(item as Record<string, unknown>)),
  };
}

export async function createCategory(payload: Record<string, unknown>) {
  const response = await apiClient.post<ApiResponse<Record<string, unknown>>>("/api/v1/categories", payload);
  return response.data;
}

export async function updateCategory(id: number, payload: Record<string, unknown>) {
  const response = await apiClient.put<ApiResponse<Record<string, unknown>>>(`/api/v1/categories/${id}`, payload);
  return response.data;
}

export async function deleteCategory(id: number) {
  const response = await apiClient.delete<ApiResponse<Record<string, unknown>>>(`/api/v1/categories/${id}`);
  return response.data;
}

export async function listBanners(params: Record<string, string | number | undefined> = { status: -1, limit: 50 }) {
  const response = await apiClient.get<ApiResponse<Array<Record<string, unknown>>>>("/api/v1/banners", {
    params,
  });
  const payload = response.data;
  return {
    ...payload,
    data: (payload.data || []).map((item) => normalizeBanner(item as Record<string, unknown>)),
  };
}

export async function createBanner(payload: Record<string, unknown>) {
  const response = await apiClient.post<ApiResponse<Record<string, unknown>>>("/api/v1/banners", payload);
  return response.data;
}

export async function updateBanner(id: number, payload: Record<string, unknown>) {
  const response = await apiClient.put<ApiResponse<Record<string, unknown>>>(`/api/v1/banners/${id}`, payload);
  return response.data;
}

export async function deleteBanner(id: number) {
  const response = await apiClient.delete<ApiResponse<Record<string, unknown>>>(`/api/v1/banners/${id}`);
  return response.data;
}

export async function listOrders(params: Record<string, string | number | undefined>) {
  const response = await apiClient.get<
    ApiResponse<{
      list: Array<{
        id: number;
        order_no: string;
        user_id: number;
        status: number;
        pay_amount: string;
        receiver_name: string;
        created_at: string;
      }>;
      total: number;
    }>
  >("/api/v1/orders", { params });
  const payload = response.data;
  return {
    ...payload,
    data: {
      ...payload.data,
      list: (payload.data?.list || []).map((item) => normalizeOrder(item as unknown as Record<string, unknown>)),
    },
  };
}

export async function listSeckillActivities(params: Record<string, string | number | boolean | undefined>) {
  const response = await apiClient.get<
    ApiResponse<{
      list: Array<{
        id: number;
        name: string;
        sku_id: number;
        sku_name: string;
        seckill_price: string;
        original_price: string;
        stock: number;
        sold: number;
        status: number;
        enable_status: number;
      }>;
      total: number;
    }>
  >("/api/v1/seckill/activities", { params });
  const payload = response.data;
  return {
    ...payload,
    data: {
      ...payload.data,
      list: (payload.data?.list || []).map((item) => normalizeSeckillActivity(item as unknown as Record<string, unknown>)),
    },
  };
}

export async function createSeckillActivity(payload: Record<string, unknown>) {
  const response = await apiClient.post<ApiResponse<Record<string, unknown>>>("/api/v1/seckill/activities", payload);
  return response.data;
}

export async function updateSeckillActivity(id: number, payload: Record<string, unknown>) {
  const response = await apiClient.put<ApiResponse<Record<string, unknown>>>(
    `/api/v1/seckill/activities/${id}`,
    payload,
  );
  return response.data;
}

export async function deleteSeckillActivity(id: number) {
  const response = await apiClient.delete<ApiResponse<Record<string, unknown>>>(`/api/v1/seckill/activities/${id}`);
  return response.data;
}
