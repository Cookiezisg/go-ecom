import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Link } from "react-router-dom";
import { getCart, removeCartItems, updateCartQuantity } from "@/api/store";
import { useAuthStore } from "@/stores/auth";

export function CartPage() {
  const queryClient = useQueryClient();
  const profile = useAuthStore((state) => state.profile);

  const cartQuery = useQuery({
    queryKey: ["cart", profile?.id],
    queryFn: () => getCart(profile!.id),
    enabled: Boolean(profile?.id),
  });

  const updateMutation = useMutation({
    mutationFn: updateCartQuantity,
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["cart", profile?.id] });
    },
  });
  const removeMutation = useMutation({
    mutationFn: (skuId: number) => removeCartItems(profile!.id, [skuId]),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["cart", profile?.id] });
    },
  });

  if (!profile) {
    return <div className="panel">请先登录后查看购物车。</div>;
  }

  const items = cartQuery.data?.data ?? [];
  const total = items.reduce((sum, item) => sum + Number(item.price) * item.quantity, 0);

  return (
    <section className="stack">
      <div className="section-head">
        <h1>购物车</h1>
        <span className="muted">已选商品会在这里统一结算。</span>
      </div>
      {cartQuery.isLoading ? <div className="panel">正在加载购物车...</div> : null}
      {cartQuery.isError ? <div className="error-box">{(cartQuery.error as Error).message}</div> : null}
      {removeMutation.isError ? <div className="error-box">{(removeMutation.error as Error).message}</div> : null}
      <div className="cart-list">
        {items.map((item) => (
          <div className="cart-row panel" key={item.id}>
            <div>
              <h3>{item.product_name}</h3>
              <p className="muted">{item.sku_name}</p>
            </div>
            <div className="cart-meta">
              <span>¥{item.price}</span>
              <span>{item.stock_status}</span>
            </div>
            <div className="quantity-box">
              <button
                className="ghost-button"
                disabled={updateMutation.isPending || removeMutation.isPending}
                onClick={() =>
                  updateMutation.mutate({
                    user_id: profile.id,
                    sku_id: item.sku_id,
                    quantity: Math.max(1, item.quantity - 1),
                  })
                }
                type="button"
              >
                -
              </button>
              <span>{item.quantity}</span>
              <button
                className="ghost-button"
                disabled={updateMutation.isPending || removeMutation.isPending}
                onClick={() =>
                  updateMutation.mutate({
                    user_id: profile.id,
                    sku_id: item.sku_id,
                    quantity: item.quantity + 1,
                  })
                }
                type="button"
              >
                +
              </button>
              <button
                className="ghost-button"
                disabled={updateMutation.isPending || removeMutation.isPending}
                onClick={() => removeMutation.mutate(item.sku_id)}
                type="button"
              >
                {removeMutation.isPending ? "删除中..." : "删除"}
              </button>
            </div>
          </div>
        ))}
        {!cartQuery.isLoading && items.length === 0 ? <div className="panel">购物车还是空的。</div> : null}
      </div>
      <div className="panel summary-row">
        <strong>合计：¥{total.toFixed(2)}</strong>
        <Link className="primary-button" to="/checkout">
          去结算
        </Link>
      </div>
    </section>
  );
}
