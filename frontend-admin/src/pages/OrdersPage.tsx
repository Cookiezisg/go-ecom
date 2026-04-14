import { useQuery } from "@tanstack/react-query";
import { listOrders } from "@/api/admin";

export function OrdersPage() {
  const query = useQuery({
    queryKey: ["admin-orders"],
    queryFn: () => listOrders({ page: 1, page_size: 20, status: -1 }),
  });

  const orders = query.data?.data?.list ?? [];

  return (
    <section className="table-card">
      <div className="card-head">
        <h2>订单管理</h2>
        <span className="muted">查看订单信息。</span>
      </div>
      {query.isLoading ? <div>加载中...</div> : null}
      {query.isError ? <div className="error-box">{(query.error as Error).message}</div> : null}
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
          {orders.map((order) => (
            <tr key={order.id}>
              <td>{order.order_no}</td>
              <td>{order.user_id}</td>
              <td>{order.status}</td>
              <td>{order.pay_amount}</td>
              <td>{order.receiver_name || "-"}</td>
              <td>{order.created_at || "-"}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </section>
  );
}
