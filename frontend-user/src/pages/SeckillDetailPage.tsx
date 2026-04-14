import { useMutation, useQuery } from "@tanstack/react-query";
import { useParams } from "react-router-dom";
import { getSeckillActivity, seckill } from "@/api/store";
import { useAuthStore } from "@/stores/auth";

export function SeckillDetailPage() {
  const { id = "" } = useParams();
  const profile = useAuthStore((state) => state.profile);
  const query = useQuery({
    queryKey: ["seckill", id],
    queryFn: () => getSeckillActivity(id),
  });

  const mutation = useMutation({
    mutationFn: () => seckill({ user_id: profile!.id, sku_id: query.data!.data.sku_id, quantity: 1 }),
  });

  const item = query.data?.data;

  return (
    <section className="stack">
      {item ? (
        <div className="detail-main panel">
          <div className="detail-image">
            {item.sku_image ? <img alt={item.name} src={item.sku_image} /> : <span>SECKILL</span>}
          </div>
          <div className="detail-content">
            <span className="eyebrow">秒杀详情</span>
            <h1>{item.name}</h1>
            <p>{item.sku_name}</p>
            <div className="detail-price">
              <strong>¥{item.seckill_price}</strong>
              <span>¥{item.original_price}</span>
            </div>
            <div className="detail-meta">
              <span>库存 {item.stock}</span>
              <span>已售 {item.sold}</span>
              <span>状态 {item.status}</span>
            </div>
            <button
              className="primary-button"
              disabled={!profile || mutation.isPending}
              onClick={() => mutation.mutate()}
              type="button"
            >
              {!profile ? "登录后抢购" : mutation.isPending ? "抢购中..." : "立即秒杀"}
            </button>
            {mutation.isSuccess ? (
              <div className="success-box">{mutation.data.data?.message || "请求已提交"}</div>
            ) : null}
          </div>
        </div>
      ) : (
        <div className="panel">正在加载秒杀活动...</div>
      )}
    </section>
  );
}
