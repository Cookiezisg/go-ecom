import { FormEvent, useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { createSku, deleteSku, listSkus, updateSku } from "@/api/admin";

type SkuForm = {
  id?: number;
  product_id: number;
  sku_code: string;
  name: string;
  price: number;
  original_price: number;
  stock: number;
  image: string;
  status: number;
};

export function SkusPage() {
  const queryClient = useQueryClient();
  const [editing, setEditing] = useState<SkuForm | null>(null);
  const query = useQuery({
    queryKey: ["admin-skus"],
    queryFn: () => listSkus({ page: 1, page_size: 20, status: -1, product_id: 0 }),
  });
  const saveMutation = useMutation({
    mutationFn: (payload: SkuForm) =>
      payload.id
        ? updateSku(payload.id, { ...payload, specs: {} })
        : createSku({ ...payload, specs: {} }),
    onSuccess: () => {
      setEditing(null);
      void queryClient.invalidateQueries({ queryKey: ["admin-skus"] });
    },
  });
  const deleteMutation = useMutation({
    mutationFn: deleteSku,
    onSuccess: () => void queryClient.invalidateQueries({ queryKey: ["admin-skus"] }),
  });

  function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const formData = new FormData(event.currentTarget);
    saveMutation.mutate({
      ...(editing ?? {}),
      product_id: Number(formData.get("product_id") || 0),
      sku_code: String(formData.get("sku_code") || ""),
      name: String(formData.get("name") || ""),
      price: Number(formData.get("price") || 0),
      original_price: Number(formData.get("original_price") || 0),
      stock: Number(formData.get("stock") || 0),
      image: String(formData.get("image") || ""),
      status: Number(formData.get("status") || 1),
    });
  }

  const list = query.data?.data?.list ?? [];

  return (
    <section className="admin-grid two-panel">
      <div className="table-card">
        <div className="card-head">
          <h2>SKU 管理</h2>
          <button className="outline-button" onClick={() => setEditing(null)} type="button">
            新建 SKU
          </button>
        </div>
        <table className="table">
          <thead>
            <tr>
              <th>ID</th>
              <th>商品 ID</th>
              <th>SKU 编码</th>
              <th>名称</th>
              <th>价格</th>
              <th>库存</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            {list.map((sku) => (
              <tr key={String(sku.id)}>
                <td>{String(sku.id)}</td>
                <td>{String(sku.product_id)}</td>
                <td>{String(sku.sku_code)}</td>
                <td>{String(sku.name)}</td>
                <td>{String(sku.price)}</td>
                <td>{String(sku.stock)}</td>
                <td>
                  <div className="action-row">
                    <button className="table-button" onClick={() => setEditing(sku as unknown as SkuForm)} type="button">
                      编辑
                    </button>
                    <button className="table-button danger" onClick={() => deleteMutation.mutate(Number(sku.id))} type="button">
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
        <h2>{editing ? "编辑 SKU" : "新建 SKU"}</h2>
        <form className="admin-form" key={editing?.id ?? "new-sku"} onSubmit={handleSubmit}>
          <input defaultValue={editing?.product_id ?? ""} name="product_id" placeholder="商品 ID" required />
          <input defaultValue={editing?.sku_code ?? ""} name="sku_code" placeholder="SKU 编码" required />
          <input defaultValue={editing?.name ?? ""} name="name" placeholder="SKU 名称" required />
          <input defaultValue={editing?.price ?? ""} name="price" placeholder="价格" required />
          <input defaultValue={editing?.original_price ?? ""} name="original_price" placeholder="原价" />
          <input defaultValue={editing?.stock ?? ""} name="stock" placeholder="库存" />
          <input defaultValue={editing?.image ?? ""} name="image" placeholder="图片 URL" />
          <select defaultValue={String(editing?.status ?? 1)} name="status">
            <option value="0">下架</option>
            <option value="1">上架</option>
          </select>
          <button className="primary-button" type="submit">
            保存 SKU
          </button>
        </form>
      </div>
    </section>
  );
}
