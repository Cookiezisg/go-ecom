import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useParams } from "react-router-dom";
import {
  confirmReceive,
  createPayment,
  getLogistics,
  getOrder,
  getPaymentStatus,
  getTracking,
} from "@/api/store";
import { useAuthStore } from "@/stores/auth";
import { getOrderStatusText } from "@/utils/order";

export function OrderDetailPage() {
  const queryClient = useQueryClient();
  const { id = "" } = useParams();
  const profile = useAuthStore((state) => state.profile)!;
  const orderQuery = useQuery({
    queryKey: ["order", id],
    queryFn: () => getOrder(id),
  });

  const paymentMutation = useMutation({
    mutationFn: () => {
      const order = orderQuery.data!.data;
      return createPayment({
        order_id: order.id,
        order_no: order.order_no,
        user_id: profile.id,
        amount: order.pay_amount || order.total_amount,
        payment_method: 1,
      });
    },
  });

  const paymentStatusQuery = useQuery({
    queryKey: ["payment", paymentMutation.data?.data?.payment_no],
    queryFn: () => getPaymentStatus(String(paymentMutation.data?.data?.payment_no)),
    enabled: Boolean(paymentMutation.data?.data?.payment_no),
  });
  const logisticsQuery = useQuery({
    queryKey: ["logistics", orderQuery.data?.data?.id],
    queryFn: () => getLogistics(orderQuery.data!.data.id),
    enabled: Boolean(orderQuery.data?.data?.id && orderQuery.data?.data?.status >= 2),
  });
  const trackingQuery = useQuery({
    queryKey: ["tracking", logisticsQuery.data?.data?.logistics_no],
    queryFn: () => getTracking(String(logisticsQuery.data?.data?.logistics_no)),
    enabled: Boolean(logisticsQuery.data?.data?.logistics_no),
  });
  const confirmMutation = useMutation({
    mutationFn: () => confirmReceive(orderQuery.data!.data.id, orderQuery.data!.data.order_no),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ["order", id] });
    },
  });

  const order = orderQuery.data?.data;

  return (
    <section className="stack">
      {order ? (
        <>
          <div className="panel">
            <div className="section-head">
              <h1>订单详情</h1>
              <span>{getOrderStatusText(order.status)}</span>
            </div>
            <div className="detail-meta wrap">
              <span>订单号：{order.order_no}</span>
              <span>下单时间：{order.created_at || "-"}</span>
              <span>收货人：{order.receiver_name || "-"}</span>
              <span>电话：{order.receiver_phone || "-"}</span>
            </div>
            <p className="muted">地址：{order.receiver_address || "暂无"}</p>
            <p className="muted">备注：{order.remark || "无"}</p>
          </div>
          <div className="panel">
            <h2>商品项</h2>
            <div className="cart-list">
              {order.items.map((item) => (
                <div className="cart-row" key={item.id}>
                  <div>
                    <strong>{item.product_name}</strong>
                    <p className="muted">{item.sku_name}</p>
                  </div>
                  <span>
                    ¥{item.price} x {item.quantity}
                  </span>
                </div>
              ))}
            </div>
          </div>
          <div className="panel summary-row">
            <strong>应付：¥{order.pay_amount || order.total_amount}</strong>
            <div className="hero-actions">
              {order.status === 1 ? (
                <button className="primary-button" onClick={() => paymentMutation.mutate()} type="button">
                  {paymentMutation.isPending ? "创建支付中..." : "创建支付单"}
                </button>
              ) : null}
              {order.status === 3 ? (
                <button className="secondary-button" onClick={() => confirmMutation.mutate()} type="button">
                  {confirmMutation.isPending ? "确认中..." : "确认收货"}
                </button>
              ) : null}
            </div>
          </div>
          {paymentMutation.data ? (
            <div className="panel">
              <h2>支付信息</h2>
              <p className="muted">支付单号：{String(paymentMutation.data.data?.payment_no || "-")}</p>
              <p className="muted">支付链接：{paymentMutation.data.pay_url || "暂无支付链接"}</p>
              {paymentStatusQuery.data ? (
                <p className="muted">支付状态：{String(paymentStatusQuery.data.data)}</p>
              ) : null}
            </div>
          ) : null}
          {logisticsQuery.data?.data ? (
            <div className="panel">
              <h2>物流信息</h2>
              <p className="muted">物流单号：{logisticsQuery.data.data.logistics_no}</p>
              <p className="muted">物流公司：{logisticsQuery.data.data.company_name || logisticsQuery.data.data.company_code || "-"}</p>
              <p className="muted">物流状态：{String(logisticsQuery.data.data.status)}</p>
              <div className="review-list">
                {(trackingQuery.data?.data ?? []).map((item, index) => (
                  <article className="review-card" key={`${item.time}-${index}`}>
                    <div className="review-head">
                      <strong>{item.status}</strong>
                      <span>{item.time}</span>
                    </div>
                    <p className="muted">
                      {item.location} {item.remark}
                    </p>
                  </article>
                ))}
                {!trackingQuery.isLoading && (trackingQuery.data?.data?.length ?? 0) === 0 ? (
                  <div className="muted">暂无物流轨迹</div>
                ) : null}
              </div>
            </div>
          ) : null}
        </>
      ) : (
        <div className="panel">正在加载订单详情...</div>
      )}
    </section>
  );
}
