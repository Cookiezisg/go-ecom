import { useQuery } from "@tanstack/react-query";
import { getUserCoupons, listCoupons } from "@/api/store";
import { useAuthStore } from "@/stores/auth";

export function CouponsPage() {
  const profile = useAuthStore((state) => state.profile)!;
  const couponsQuery = useQuery({
    queryKey: ["all-coupons"],
    queryFn: listCoupons,
  });
  const userCouponsQuery = useQuery({
    queryKey: ["user-coupons", profile.id],
    queryFn: () => getUserCoupons(profile.id, 0),
  });

  return (
    <section className="stack two-columns">
      <div className="panel">
        <h1>平台优惠券</h1>
        <div className="coupon-list">
          {(couponsQuery.data?.data ?? []).map((coupon) => (
            <article className="coupon-card" key={coupon.id}>
              <strong>{coupon.name}</strong>
              <p className="muted">优惠值：{coupon.discount_value}</p>
              <p className="muted">有效期：{coupon.valid_end_time || "-"}</p>
            </article>
          ))}
        </div>
      </div>
      <div className="panel">
        <h2>我的优惠券</h2>
        <div className="coupon-list">
          {(userCouponsQuery.data?.data ?? []).map((coupon) => (
            <article className="coupon-card" key={coupon.id}>
              <strong>用户券 #{coupon.id}</strong>
              <p className="muted">状态：{coupon.status}</p>
              <p className="muted">到期时间：{coupon.expire_at || "-"}</p>
            </article>
          ))}
          {!userCouponsQuery.isLoading && (userCouponsQuery.data?.data?.length ?? 0) === 0 ? (
            <div className="muted">暂无可用优惠券</div>
          ) : null}
        </div>
      </div>
    </section>
  );
}
