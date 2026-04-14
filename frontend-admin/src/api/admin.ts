import { apiClient } from "@/api/client";

export interface ApiResponse<T> {
  code: number;
  message: string;
  data: T;
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
  return response.data;
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
  return response.data;
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
  return response.data;
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
  return response.data;
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
  return response.data;
}

export async function listCategoryTree(status = 1) {
  const response = await apiClient.get<ApiResponse<Array<Record<string, unknown>>>>("/api/v1/categories/tree", {
    params: { status },
  });
  return response.data;
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
  return response.data;
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
  return response.data;
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
  return response.data;
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
