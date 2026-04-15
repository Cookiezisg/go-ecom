import { FormEvent, useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { createProduct, deleteProduct, listProducts, updateProduct } from "@/api/admin";
import { DataTableControls } from "@/components/DataTableControls";
import { uploadImage } from "@/api/upload";

type ProductForm = {
  id?: number;
  name: string;
  subtitle: string;
  category_id: number;
  brand_id: number;
  main_image: string;
  local_main_image: string;
  detail: string;
  price: number;
  original_price: number;
  stock: number;
  status: number;
  is_hot: number;
};

const emptyForm: ProductForm = {
  name: "",
  subtitle: "",
  category_id: 0,
  brand_id: 0,
  main_image: "",
  local_main_image: "",
  detail: "",
  price: 0,
  original_price: 0,
  stock: 0,
  status: 1,
  is_hot: 0,
};

function toProductForm(input: Record<string, unknown>): ProductForm {
  return {
    id: Number(input.id || 0),
    name: String(input.name || ""),
    subtitle: String(input.subtitle || ""),
    category_id: Number(input.category_id || 0),
    brand_id: Number(input.brand_id || 0),
    main_image: String(input.main_image || ""),
    local_main_image: String(input.local_main_image || ""),
    detail: String(input.detail || ""),
    price: Number(input.price || 0),
    original_price: Number(input.original_price || 0),
    stock: Number(input.stock || 0),
    status: Number(input.status || 1),
    is_hot: Number(input.is_hot || 0),
  };
}

export function ProductsAdminPage() {
  const queryClient = useQueryClient();
  const [editing, setEditing] = useState<ProductForm | null>(null);
  const [keyword, setKeyword] = useState("");
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const query = useQuery({
    queryKey: ["admin-products", page, pageSize, keyword],
    queryFn: () => listProducts({ page, page_size: pageSize, status: 1, keyword }),
  });

  const saveMutation = useMutation({
    mutationFn: async (payload: ProductForm) =>
      payload.id ? updateProduct(payload.id, payload) : createProduct(payload),
    onSuccess: () => {
      setEditing(null);
      void queryClient.invalidateQueries({ queryKey: ["admin-products"] });
    },
  });
  const deleteMutation = useMutation({
    mutationFn: deleteProduct,
    onSuccess: () => void queryClient.invalidateQueries({ queryKey: ["admin-products"] }),
  });
  const uploadMutation = useMutation({
    mutationFn: (file: File) => uploadImage(file, "image"),
  });

  const products = query.data?.data?.list ?? [];

  function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const formData = new FormData(event.currentTarget);
    saveMutation.mutate({
      ...(editing ?? {}),
      name: String(formData.get("name") || ""),
      subtitle: String(formData.get("subtitle") || ""),
      category_id: Number(formData.get("category_id") || 0),
      brand_id: Number(formData.get("brand_id") || 0),
      main_image: String(formData.get("main_image") || ""),
      local_main_image: String(formData.get("local_main_image") || ""),
      detail: String(formData.get("detail") || ""),
      price: Number(formData.get("price") || 0),
      original_price: Number(formData.get("original_price") || 0),
      stock: Number(formData.get("stock") || 0),
      status: Number(formData.get("status") || 1),
      is_hot: Number(formData.get("is_hot") || 0),
    });
  }

  return (
    <section className="admin-grid two-panel">
      <div className="table-card">
        <div className="card-head">
          <h2>商品管理</h2>
          <button className="outline-button" onClick={() => setEditing(null)} type="button">
            新建商品
          </button>
        </div>
        <DataTableControls
          onPageChange={setPage}
          onPageSizeChange={(size) => {
            setPageSize(size);
            setPage(1);
          }}
          onSearchChange={(value) => {
            setKeyword(value);
            setPage(1);
          }}
          page={page}
          pageSize={pageSize}
          searchPlaceholder="搜索商品名、副标题、分类 ID"
          searchValue={keyword}
          total={query.data?.data?.total ?? products.length}
        />
        <table className="table">
          <thead>
            <tr>
              <th>ID</th>
              <th>商品名</th>
              <th>价格</th>
              <th>库存</th>
              <th>销量</th>
              <th>状态</th>
              <th>操作</th>
            </tr>
          </thead>
        <tbody>
            {products.map((product) => (
              <tr key={product.id}>
                <td>{product.id}</td>
                <td>{product.name}</td>
                <td>¥{product.price}</td>
                <td>{product.stock}</td>
                <td>{product.sales}</td>
                <td>{product.status}</td>
                <td>
                  <div className="action-row">
                    <button className="table-button" onClick={() => setEditing(toProductForm(product as unknown as Record<string, unknown>))} type="button">
                      编辑
                    </button>
                    <button className="table-button danger" onClick={() => deleteMutation.mutate(product.id)} type="button">
                      删除
                    </button>
                  </div>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
      <div className="table-card">
        <h2>{editing ? "编辑商品" : "新建商品"}</h2>
        <form className="admin-form" key={editing?.id ?? "new"} onSubmit={handleSubmit}>
          <input defaultValue={editing?.name ?? emptyForm.name} name="name" placeholder="商品名" required />
          <input defaultValue={editing?.subtitle ?? emptyForm.subtitle} name="subtitle" placeholder="副标题" />
          <input defaultValue={editing?.category_id ?? emptyForm.category_id} name="category_id" placeholder="分类 ID" />
          <input defaultValue={editing?.brand_id ?? emptyForm.brand_id} name="brand_id" placeholder="品牌 ID" />
          <input defaultValue={editing?.price ?? emptyForm.price} name="price" placeholder="价格" />
          <input defaultValue={editing?.original_price ?? emptyForm.original_price} name="original_price" placeholder="原价" />
          <input defaultValue={editing?.stock ?? emptyForm.stock} name="stock" placeholder="库存" />
          <input defaultValue={editing?.main_image ?? emptyForm.main_image} name="main_image" placeholder="主图 URL" />
          <input
            accept="image/*"
            onChange={(event) => {
              const file = event.target.files?.[0];
              if (!file) return;
              uploadMutation.mutate(file, {
                onSuccess: (response) => {
                  const target = document.querySelector<HTMLInputElement>('input[name="main_image"]');
                  if (target && response.data?.file_url) {
                    target.value = response.data.file_url;
                  }
                },
              });
            }}
            type="file"
          />
          <textarea className="admin-text-area" defaultValue={editing?.detail ?? emptyForm.detail} name="detail" placeholder="详情" />
          <select defaultValue={String(editing?.status ?? emptyForm.status)} name="status">
            <option value="0">下架</option>
            <option value="1">上架</option>
            <option value="2">待审核</option>
          </select>
          <select defaultValue={String(editing?.is_hot ?? emptyForm.is_hot)} name="is_hot">
            <option value="0">普通</option>
            <option value="1">热门</option>
          </select>
          <button className="primary-button" disabled={saveMutation.isPending} type="submit">
            {saveMutation.isPending ? "保存中..." : "保存商品"}
          </button>
        </form>
      </div>
    </section>
  );
}
