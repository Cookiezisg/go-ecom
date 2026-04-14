import { useQueries } from "@tanstack/react-query";
import { listOrders, listProducts, listSeckillActivities, listUsers } from "@/api/admin";

export function DashboardPage() {
  const results = useQueries({
    queries: [
      { queryKey: ["dashboard-users"], queryFn: () => listUsers({ page: 1, page_size: 1 }) },
      { queryKey: ["dashboard-products"], queryFn: () => listProducts({ page: 1, page_size: 1 }) },
      { queryKey: ["dashboard-orders"], queryFn: () => listOrders({ page: 1, page_size: 1, status: -1 }) },
      {
        queryKey: ["dashboard-seckill"],
        queryFn: () => listSeckillActivities({ page: 1, page_size: 1, include_disabled: true }),
      },
    ],
  });

  const [users, products, orders, seckill] = results;

  return (
    <section className="admin-grid">
      <StatCard label="用户总数" value={users.data?.data?.total ?? "--"} />
      <StatCard label="商品总数" value={products.data?.data?.total ?? "--"} />
      <StatCard label="订单总数" value={orders.data?.data?.total ?? "--"} />
      <StatCard label="秒杀活动数" value={seckill.data?.data?.total ?? "--"} />

      <div className="table-card span-2">
        <h2>当前说明</h2>
        <p>
          当前仪表盘优先展示基础聚合卡片。由于后端暂时没有专门的运营聚合接口，这里先使用现有列表接口做轻量统计。
        </p>
      </div>
    </section>
  );
}

function StatCard({ label, value }: { label: string; value: string | number }) {
  return (
    <article className="stat-card">
      <span>{label}</span>
      <strong>{value}</strong>
    </article>
  );
}
