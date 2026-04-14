import { useMemo, useState } from "react";
import { useMutation, useQuery } from "@tanstack/react-query";
import { Link, useNavigate } from "react-router-dom";
import { createOrder, getAddresses, getCart, listCoupons } from "@/api/store";
import { useAuthStore } from "@/stores/auth";

export function CheckoutPage() {
  const navigate = useNavigate();
  const profile = useAuthStore((state) => state.profile)!;
  const [selectedAddress, setSelectedAddress] = useState<number>(0);
  const [selectedCoupon, setSelectedCoupon] = useState<number>(0);
  const [remark, setRemark] = useState("");

  const cartQuery = useQuery({
    queryKey: ["checkout-cart", profile.id],
    queryFn: () => getCart(profile.id),
  });

  const addressQuery = useQuery({
    queryKey: ["checkout-address", profile.id],
    queryFn: () => getAddresses(profile.id),
  });

  const couponQuery = useQuery({
    queryKey: ["checkout-coupons"],
    queryFn: listCoupons,
  });

  const createMutation = useMutation({
    mutationFn: () =>
      createOrder({
        user_id: profile.id,
        address_id: selectedAddress,
        items: (cartQuery.data?.data ?? []).map((item) => ({ sku_id: item.sku_id, quantity: item.quantity })),
        order_type: 1,
        coupon_id: selectedCoupon || undefined,
        remark,
      }),
    onSuccess: (response) => {
      navigate(`/orders/${response.data.id}`);
    },
  });

  const items = cartQuery.data?.data ?? [];
  const total = useMemo(
    () => items.reduce((sum, item) => sum + Number(item.price) * item.quantity, 0),
    [items],
  );
  const addresses = addressQuery.data?.data ?? [];

  return (
    <section className="stack">
      <div className="section-head">
        <h1>确认订单</h1>
        <span className="muted">请确认收货信息和商品清单。</span>
      </div>
      <div className="panel">
        <h2>收货地址</h2>
        {addresses.length === 0 ? (
          <div className="muted">
            暂无地址，先去 <Link to="/addresses">地址管理</Link> 新增。
          </div>
        ) : (
          <div className="address-list">
            {addresses.map((address) => (
              <label className="address-card selectable" key={address.id}>
                <input
                  checked={selectedAddress === address.id}
                  name="address_id"
                  onChange={() => setSelectedAddress(address.id)}
                  type="radio"
                />
                <strong>
                  {address.receiver_name} · {address.receiver_phone}
                </strong>
                <p className="muted">
                  {address.province} {address.city} {address.district} {address.detail}
                </p>
              </label>
            ))}
          </div>
        )}
      </div>
      <div className="panel">
        <h2>商品清单</h2>
        <div className="cart-list">
          {items.map((item) => (
            <div className="cart-row" key={item.id}>
              <div>
                <strong>{item.product_name}</strong>
                <p className="muted">{item.sku_name}</p>
              </div>
              <div className="cart-meta">
                <span>¥{item.price}</span>
                <span>x {item.quantity}</span>
              </div>
            </div>
          ))}
        </div>
      </div>
      <div className="stack two-columns">
        <div className="panel">
          <h2>优惠券</h2>
          <select onChange={(e) => setSelectedCoupon(Number(e.target.value))} value={selectedCoupon}>
            <option value="0">不使用优惠券</option>
            {(couponQuery.data?.data ?? []).map((coupon) => (
              <option key={coupon.id} value={coupon.id}>
                {coupon.name}
              </option>
            ))}
          </select>
        </div>
        <div className="panel">
          <h2>订单备注</h2>
          <textarea
            className="text-area"
            onChange={(e) => setRemark(e.target.value)}
            placeholder="给商家留言"
            value={remark}
          />
        </div>
      </div>
      <div className="panel summary-row">
        <strong>应付金额：¥{total.toFixed(2)}</strong>
        <button
          className="primary-button"
          disabled={!selectedAddress || items.length === 0 || createMutation.isPending}
          onClick={() => createMutation.mutate()}
          type="button"
        >
          {createMutation.isPending ? "提交中..." : "提交订单"}
        </button>
      </div>
    </section>
  );
}
