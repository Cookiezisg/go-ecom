import { useMemo, useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Link } from "react-router-dom";
import { cancelOrder, listOrders } from "@/api/store";
import { useAuthStore } from "@/stores/auth";
import { getOrderStatusText } from "@/utils/order";

const tabs = [
  { label: "全部", value: -1 },
  { label: "待支付", value: 1 },
  { label: "待发货", value: 2 },
  { label: "待收货", value: 3 },
  { label: "已完成", value: 4 },
];

export function OrdersPage() {
  const queryClient = useQueryClient();
  const profile = useAuthStore((state) => state.profile)!;
  const [status, setStatus] = useState(-1);

  const query = useQuery({
    queryKey: ["orders", profile.id, status],
    queryFn: () => listOrders(profile.id, status),
  });

  const cancelMutation = useMutation({
    mutationFn: (id: number) => cancelOrder(id, "用户取消"),
    onSuccess: () => void queryClient.invalidateQueries({ queryKey: ["orders", profile.id] }),
  });

  const orders = useMemo(() => query.data?.data?.list ?? [], [query.data]);

  return (
    <section className="stack">
      <div className="section-head">
        <h1>我的订单</h1>
        <span className="muted">查看订单状态和物流信息。</span>
      </div>
      <div className="filter-chips">
        {tabs.map((tab) => (
          <button
            className={status === tab.value ? "tab-button active" : "tab-button"}
            key={tab.value}
            onClick={() => setStatus(tab.value)}
            type="button"
          >
            {tab.label}
          </button>
        ))}
      </div>
      <div className="order-list">
        {orders.map((order) => (
          <article className="panel" key={order.id}>
            <div className="review-head">
              <strong>{order.order_no}</strong>
              <span>{getOrderStatusText(order.status)}</span>
            </div>
            <div className="stack compact">
              {(order.items || []).map((item) => (
                <div className="cart-row" key={item.id}>
                  <span>{item.product_name}</span>
                  <span>
                    ¥{item.price} x {item.quantity}
                  </span>
                </div>
              ))}
            </div>
            <div className="summary-row">
              <strong>实付：¥{order.pay_amount || order.total_amount}</strong>
              <div className="hero-actions">
                <Link className="secondary-button" to={`/orders/${order.id}`}>
                  查看详情
                </Link>
                {order.status === 1 ? (
                  <button
                    className="ghost-button"
                    onClick={() => cancelMutation.mutate(order.id)}
                    type="button"
                  >
                    取消订单
                  </button>
                ) : null}
              </div>
            </div>
          </article>
        ))}
        {!query.isLoading && orders.length === 0 ? <div className="panel">暂无订单</div> : null}
      </div>
    </section>
  );
}
