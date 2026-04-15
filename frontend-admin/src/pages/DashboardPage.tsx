import { useMemo } from "react";
import { useQuery } from "@tanstack/react-query";
import {
  listBanners,
  listCategories,
  listOrders,
  listProducts,
  listSeckillActivities,
  listSkus,
  listUsers,
} from "@/api/admin";

export function DashboardPage() {
  const users = useQuery({
    queryKey: ["dashboard-users"],
    queryFn: () => listUsers({ page: 1, page_size: 1, status: 0 }),
  });
  const productsAll = useQuery({
    queryKey: ["dashboard-products-all"],
    queryFn: () => listProducts({ page: 1, page_size: 1, status: -1 }),
  });
  const productsOnline = useQuery({
    queryKey: ["dashboard-products-online"],
    queryFn: () => listProducts({ page: 1, page_size: 1, status: 1 }),
  });
  const productsHot = useQuery({
    queryKey: ["dashboard-products-hot"],
    queryFn: () => listProducts({ page: 1, page_size: 1, status: -1, is_hot: 1 }),
  });
  const skusAll = useQuery({
    queryKey: ["dashboard-skus-all"],
    queryFn: () => listSkus({ page: 1, page_size: 1, status: -1, product_id: 0 }),
  });
  const skusOnline = useQuery({
    queryKey: ["dashboard-skus-online"],
    queryFn: () => listSkus({ page: 1, page_size: 1, status: 1, product_id: 0 }),
  });
  const skusLowStock = useQuery({
    queryKey: ["dashboard-skus-low-stock"],
    queryFn: () => listSkus({ page: 1, page_size: 100, status: 1, product_id: 0 }),
  });
  const ordersAll = useQuery({
    queryKey: ["dashboard-orders-all"],
    queryFn: () => listOrders({ page: 1, page_size: 1, status: -1 }),
  });
  const ordersPendingPay = useQuery({
    queryKey: ["dashboard-orders-pending-pay"],
    queryFn: () => listOrders({ page: 1, page_size: 1, status: 1 }),
  });
  const ordersPendingShip = useQuery({
    queryKey: ["dashboard-orders-pending-ship"],
    queryFn: () => listOrders({ page: 1, page_size: 1, status: 2 }),
  });
  const ordersPendingReceive = useQuery({
    queryKey: ["dashboard-orders-pending-receive"],
    queryFn: () => listOrders({ page: 1, page_size: 1, status: 3 }),
  });
  const ordersCompleted = useQuery({
    queryKey: ["dashboard-orders-completed"],
    queryFn: () => listOrders({ page: 1, page_size: 1, status: 4 }),
  });
  const ordersCancelled = useQuery({
    queryKey: ["dashboard-orders-cancelled"],
    queryFn: () => listOrders({ page: 1, page_size: 1, status: 0 }),
  });
  const ordersRefunded = useQuery({
    queryKey: ["dashboard-orders-refunded"],
    queryFn: () => listOrders({ page: 1, page_size: 1, status: 5 }),
  });
  const recentOrders = useQuery({
    queryKey: ["dashboard-orders-recent"],
    queryFn: () => listOrders({ page: 1, page_size: 5, status: -1 }),
  });
  const seckillAll = useQuery({
    queryKey: ["dashboard-seckill-all"],
    queryFn: () => listSeckillActivities({ page: 1, page_size: 100, include_disabled: true }),
  });
  const seckillActive = useQuery({
    queryKey: ["dashboard-seckill-active"],
    queryFn: () => listSeckillActivities({ page: 1, page_size: 1, status: 1, include_disabled: true }),
  });
  const seckillUpcoming = useQuery({
    queryKey: ["dashboard-seckill-upcoming"],
    queryFn: () => listSeckillActivities({ page: 1, page_size: 1, status: 0, include_disabled: true }),
  });
  const seckillEnded = useQuery({
    queryKey: ["dashboard-seckill-ended"],
    queryFn: () => listSeckillActivities({ page: 1, page_size: 1, status: 2, include_disabled: true }),
  });
  const banners = useQuery({
    queryKey: ["dashboard-banners"],
    queryFn: () => listBanners({ status: -1, limit: 200 }),
  });
  const categories = useQuery({
    queryKey: ["dashboard-categories"],
    queryFn: () => listCategories({ status: 1 }),
  });

  const lowStockSkuCount = useMemo(
    () => (skusLowStock.data?.data?.list ?? []).filter((item) => item.stock > 0 && item.stock <= 10).length,
    [skusLowStock.data],
  );
  const seckillStats = useMemo(() => {
    const list = seckillAll.data?.data?.list ?? [];
    return list.reduce(
      (acc, item) => {
        acc.stock += Number(item.stock || 0);
        acc.sold += Number(item.sold || 0);
        if (item.enable_status === 0) {
          acc.disabled += 1;
        }
        return acc;
      },
      { stock: 0, sold: 0, disabled: 0 },
    );
  }, [seckillAll.data]);
  const bannerStats = useMemo(() => {
    const list = banners.data?.data ?? [];
    return list.reduce(
      (acc, item) => {
        if (item.status === 1) acc.enabled += 1;
        if (item.status === 0) acc.disabled += 1;
        return acc;
      },
      { enabled: 0, disabled: 0 },
    );
  }, [banners.data]);
  const categoryStats = useMemo(() => {
    const list = categories.data?.data ?? [];
    return list.reduce<{ level1: number; level2: number }>(
      (acc, item) => {
        if (Number(item.level) === 1) acc.level1 += 1;
        if (Number(item.level) === 2) acc.level2 += 1;
        return acc;
      },
      { level1: 0, level2: 0 },
    );
  }, [categories.data]);
  const hasError = [
    users,
    productsAll,
    productsOnline,
    productsHot,
    skusAll,
    skusOnline,
    skusLowStock,
    ordersAll,
    ordersPendingPay,
    ordersPendingShip,
    ordersPendingReceive,
    ordersCompleted,
    ordersCancelled,
    ordersRefunded,
    recentOrders,
    seckillAll,
    seckillActive,
    seckillUpcoming,
    seckillEnded,
    banners,
    categories,
  ].some((item) => item.isError);

  return (
    <section className="dashboard-stack">
      <div className="dashboard-hero table-card">
        <div>
          <h2>运营概览</h2>
          <p className="muted">
            当前仪表盘使用现有列表接口做轻量聚合，先覆盖订单、商品、秒杀和内容运营的核心状态。
          </p>
        </div>
      </div>

      {hasError ? <div className="error-box">部分统计加载失败，请检查后端服务和网关状态。</div> : null}

      <section className="admin-grid">
        <StatCard label="用户总数" value={users.data?.data?.total ?? "--"} />
        <StatCard label="商品总数" value={productsAll.data?.data?.total ?? "--"} />
        <StatCard label="订单总数" value={ordersAll.data?.data?.total ?? "--"} />
        <StatCard label="进行中秒杀" value={seckillActive.data?.data?.total ?? "--"} />
        <StatCard label="待支付订单" value={ordersPendingPay.data?.data?.total ?? "--"} />
        <StatCard label="待发货订单" value={ordersPendingShip.data?.data?.total ?? "--"} />
      </section>

      <section className="admin-grid two-panel">
        <div className="table-card">
          <div className="card-head">
            <h2>订单状态</h2>
            <span className="muted">订单履约当前分布</span>
          </div>
          <div className="metric-list">
            <MetricRow label="待支付" value={ordersPendingPay.data?.data?.total ?? 0} />
            <MetricRow label="待发货" value={ordersPendingShip.data?.data?.total ?? 0} />
            <MetricRow label="待收货" value={ordersPendingReceive.data?.data?.total ?? 0} />
            <MetricRow label="已完成" value={ordersCompleted.data?.data?.total ?? 0} />
            <MetricRow label="已取消" value={ordersCancelled.data?.data?.total ?? 0} />
            <MetricRow label="已退款" value={ordersRefunded.data?.data?.total ?? 0} />
          </div>
        </div>

        <div className="table-card">
          <div className="card-head">
            <h2>商品与库存</h2>
            <span className="muted">商品池与 SKU 规模</span>
          </div>
          <div className="metric-list">
            <MetricRow label="上架商品" value={productsOnline.data?.data?.total ?? 0} />
            <MetricRow label="热门商品" value={productsHot.data?.data?.total ?? 0} />
            <MetricRow label="SKU 总数" value={skusAll.data?.data?.total ?? 0} />
            <MetricRow label="上架 SKU" value={skusOnline.data?.data?.total ?? 0} />
            <MetricRow label="低库存 SKU" value={lowStockSkuCount} />
          </div>
        </div>
      </section>

      <section className="admin-grid two-panel">
        <div className="table-card">
          <div className="card-head">
            <h2>活动与内容</h2>
            <span className="muted">秒杀、Banner、分类概览</span>
          </div>
          <div className="metric-list">
            <MetricRow label="未开始秒杀" value={seckillUpcoming.data?.data?.total ?? 0} />
            <MetricRow label="进行中秒杀" value={seckillActive.data?.data?.total ?? 0} />
            <MetricRow label="已结束秒杀" value={seckillEnded.data?.data?.total ?? 0} />
            <MetricRow label="禁用秒杀" value={seckillStats.disabled} />
            <MetricRow label="启用 Banner" value={bannerStats.enabled} />
            <MetricRow label="禁用 Banner" value={bannerStats.disabled} />
            <MetricRow label="一级分类" value={Number(categoryStats.level1)} />
            <MetricRow label="二级分类" value={Number(categoryStats.level2)} />
          </div>
        </div>

        <div className="table-card">
          <div className="card-head">
            <h2>秒杀库存</h2>
            <span className="muted">当前活动库存与已售</span>
          </div>
          <div className="metric-list">
            <MetricRow label="秒杀活动数" value={seckillAll.data?.data?.total ?? 0} />
            <MetricRow label="总秒杀库存" value={seckillStats.stock} />
            <MetricRow label="总已售数量" value={seckillStats.sold} />
          </div>
        </div>
      </section>

      <div className="table-card">
        <div className="card-head">
          <h2>最近订单</h2>
          <span className="muted">最近创建的订单记录</span>
        </div>
        <table className="table">
          <thead>
            <tr>
              <th>订单号</th>
              <th>用户</th>
              <th>状态</th>
              <th>实付金额</th>
              <th>收货人</th>
              <th>创建时间</th>
            </tr>
          </thead>
          <tbody>
            {(recentOrders.data?.data?.list ?? []).map((order) => (
              <tr key={order.id}>
                <td>{order.order_no}</td>
                <td>{order.user_id}</td>
                <td>{getOrderStatusText(order.status)}</td>
                <td>{order.pay_amount}</td>
                <td>{order.receiver_name || "-"}</td>
                <td>{order.created_at || "-"}</td>
              </tr>
            ))}
            {!recentOrders.isLoading && (recentOrders.data?.data?.list?.length ?? 0) === 0 ? (
              <tr>
                <td className="dashboard-empty" colSpan={6}>
                  暂无订单数据
                </td>
              </tr>
            ) : null}
          </tbody>
        </table>
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

function MetricRow({ label, value }: { label: string; value: string | number }) {
  return (
    <div className="metric-row">
      <span>{label}</span>
      <strong>{value}</strong>
    </div>
  );
}

function getOrderStatusText(status: number) {
  switch (status) {
    case 0:
      return "已取消";
    case 1:
      return "待支付";
    case 2:
      return "待发货";
    case 3:
      return "待收货";
    case 4:
      return "已完成";
    case 5:
      return "已退款";
    default:
      return String(status);
  }
}
