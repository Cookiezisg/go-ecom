import { FormEvent, useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  createSeckillActivity,
  deleteSeckillActivity,
  listSeckillActivities,
  updateSeckillActivity,
} from "@/api/admin";

type SeckillForm = {
  id?: number;
  name: string;
  sku_id: number;
  seckill_price: string;
  stock: number;
  start_time: number;
  end_time: number;
  enable_status: number;
};

export function SeckillPage() {
  const queryClient = useQueryClient();
  const [editing, setEditing] = useState<SeckillForm | null>(null);
  const query = useQuery({
    queryKey: ["admin-seckill"],
    queryFn: () => listSeckillActivities({ page: 1, page_size: 20, include_disabled: true }),
  });
  const saveMutation = useMutation({
    mutationFn: (payload: SeckillForm) =>
      payload.id ? updateSeckillActivity(payload.id, payload) : createSeckillActivity(payload),
    onSuccess: () => {
      setEditing(null);
      void queryClient.invalidateQueries({ queryKey: ["admin-seckill"] });
    },
  });
  const deleteMutation = useMutation({
    mutationFn: deleteSeckillActivity,
    onSuccess: () => void queryClient.invalidateQueries({ queryKey: ["admin-seckill"] }),
  });

  function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const formData = new FormData(event.currentTarget);
    saveMutation.mutate({
      ...(editing ?? {}),
      name: String(formData.get("name") || ""),
      sku_id: Number(formData.get("sku_id") || 0),
      seckill_price: String(formData.get("seckill_price") || "0"),
      stock: Number(formData.get("stock") || 0),
      start_time: Number(formData.get("start_time") || 0),
      end_time: Number(formData.get("end_time") || 0),
      enable_status: Number(formData.get("enable_status") || 1),
    });
  }

  const activities = query.data?.data?.list ?? [];

  return (
    <section className="admin-grid two-panel">
      <div className="table-card">
        <div className="card-head">
          <h2>秒杀活动管理</h2>
          <button className="outline-button" onClick={() => setEditing(null)} type="button">
            新建活动
          </button>
        </div>
        <table className="table">
          <thead>
            <tr>
              <th>ID</th>
              <th>活动名</th>
              <th>SKU</th>
              <th>秒杀价</th>
              <th>库存</th>
              <th>状态</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            {activities.map((item) => (
              <tr key={item.id}>
                <td>{item.id}</td>
                <td>{item.name}</td>
                <td>{item.sku_name || item.sku_id}</td>
                <td>{item.seckill_price}</td>
                <td>{item.stock}</td>
                <td>
                  {item.status} / {item.enable_status}
                </td>
                <td>
                  <div className="action-row">
                    <button className="table-button" onClick={() => setEditing(item as unknown as SeckillForm)} type="button">
                      编辑
                    </button>
                    <button className="table-button danger" onClick={() => deleteMutation.mutate(item.id)} type="button">
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
        <h2>{editing ? "编辑秒杀活动" : "新建秒杀活动"}</h2>
        <form className="admin-form" key={editing?.id ?? "new-seckill"} onSubmit={handleSubmit}>
          <input defaultValue={editing?.name ?? ""} name="name" placeholder="活动名" required />
          <input defaultValue={editing?.sku_id ?? ""} name="sku_id" placeholder="SKU ID" required />
          <input defaultValue={editing?.seckill_price ?? ""} name="seckill_price" placeholder="秒杀价" required />
          <input defaultValue={editing?.stock ?? ""} name="stock" placeholder="库存" required />
          <input defaultValue={editing?.start_time ?? ""} name="start_time" placeholder="开始时间戳" required />
          <input defaultValue={editing?.end_time ?? ""} name="end_time" placeholder="结束时间戳" required />
          <select defaultValue={String(editing?.enable_status ?? 1)} name="enable_status">
            <option value="0">禁用</option>
            <option value="1">启用</option>
          </select>
          <button className="primary-button" type="submit">
            保存活动
          </button>
        </form>
      </div>
    </section>
  );
}
