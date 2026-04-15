import { FormEvent, useEffect, useMemo, useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useParams } from "react-router-dom";
import {
  addCartItem,
  createReview,
  getProduct,
  getProductReviews,
  getReviewStats,
  listSkus,
} from "@/api/store";
import { useAuthStore } from "@/stores/auth";

export function ProductDetailPage() {
  const queryClient = useQueryClient();
  const { id = "" } = useParams();
  const profile = useAuthStore((state) => state.profile);
  const [reviewMessage, setReviewMessage] = useState("");
  const [selectedSkuId, setSelectedSkuId] = useState<number>(0);

  const productQuery = useQuery({
    queryKey: ["product", id],
    queryFn: () => getProduct(id),
    enabled: Boolean(id),
  });

  const reviewsQuery = useQuery({
    queryKey: ["product-reviews", id],
    queryFn: () => getProductReviews(id),
    enabled: Boolean(id),
  });

  const skusQuery = useQuery({
    queryKey: ["product-skus", id],
    queryFn: () => listSkus({ product_id: id, status: 1, page: 1, page_size: 50 }),
    enabled: Boolean(id),
  });

  const statsQuery = useQuery({
    queryKey: ["product-review-stats", id],
    queryFn: () => getReviewStats(id),
    enabled: Boolean(id),
  });

  const addCartMutation = useMutation({
    mutationFn: (skuId: number) =>
      addCartItem({
        user_id: profile?.id || 0,
        sku_id: skuId,
        quantity: 1,
      }),
  });
  const reviewMutation = useMutation({
    mutationFn: createReview,
    onSuccess: () => {
      setReviewMessage("评价已提交");
      void queryClient.invalidateQueries({ queryKey: ["product-reviews", id] });
      void queryClient.invalidateQueries({ queryKey: ["product-review-stats", id] });
    },
  });

  const product = productQuery.data?.data;
  const skuList = skusQuery.data?.data?.list ?? [];
  const selectedSku =
    skuList.find((item) => item.id === selectedSkuId) ??
    skuList.find((item) => item.status === 1) ??
    skuList[0];
  const selectedSkuSpecs = useMemo(
    () => Object.entries(selectedSku?.specs ?? {}).filter(([, value]) => value),
    [selectedSku],
  );
  const productImage =
    selectedSku?.image || product?.local_main_image || product?.main_image || "";

  useEffect(() => {
    if (!skuList.length) {
      setSelectedSkuId(0);
      return;
    }
    setSelectedSkuId((current) => {
      if (current > 0 && skuList.some((item) => item.id === current)) {
        return current;
      }
      return skuList.find((item) => item.status === 1)?.id ?? skuList[0].id;
    });
  }, [skuList]);

  return (
    <section className="detail-shell">
      <div className="detail-main panel">
        {productQuery.isLoading ? <div>正在加载商品详情...</div> : null}
        {productQuery.isError ? (
          <div className="error-box">{(productQuery.error as Error).message}</div>
        ) : null}
        {skusQuery.isError ? (
          <div className="error-box">{(skusQuery.error as Error).message}</div>
        ) : null}
        {product ? (
          <>
            <div className="detail-image">
              {productImage ? (
                <img alt={product.name} src={productImage} />
              ) : (
                <span>NO IMAGE</span>
              )}
            </div>
            <div className="detail-content">
              <span className="eyebrow">商品详情</span>
              <h1>{product.name}</h1>
              <p>{product.subtitle || "查看商品信息、价格和用户评价。"}</p>
              <div className="detail-price">
                <strong>¥{selectedSku?.price ?? product.price}</strong>
                {selectedSku?.original_price || product.original_price ? (
                  <span>¥{selectedSku?.original_price ?? product.original_price}</span>
                ) : null}
              </div>
              <div className="detail-meta">
                <span>库存 {selectedSku?.stock ?? product.stock}</span>
                <span>销量 {product.sales}</span>
                <span>评分 {statsQuery.data?.data?.average_rating ?? 0}</span>
              </div>
              {skuList.length > 0 ? (
                <div className="sku-section">
                  <div className="sku-label">规格</div>
                  <div className="sku-options">
                    {skuList.map((sku) => (
                      <button
                        className={`tab-button ${selectedSku?.id === sku.id ? "active" : ""}`}
                        key={sku.id}
                        onClick={() => setSelectedSkuId(sku.id)}
                        type="button"
                      >
                        {sku.name}
                      </button>
                    ))}
                  </div>
                  {selectedSku ? <div className="muted">当前规格：{selectedSku.name}</div> : null}
                  {selectedSkuSpecs.length > 0 ? (
                    <div className="sku-specs">
                      {selectedSkuSpecs.map(([key, value]) => (
                        <span className="sku-spec-chip" key={`${key}-${value}`}>
                          {key}：{value}
                        </span>
                      ))}
                    </div>
                  ) : null}
                </div>
              ) : null}
              <div className="detail-actions">
                <button
                  className="primary-button"
                  disabled={!profile || addCartMutation.isPending || !selectedSku}
                  onClick={() => selectedSku && addCartMutation.mutate(selectedSku.id)}
                  type="button"
                >
                  {!profile
                    ? "登录后加入购物车"
                    : !selectedSku
                      ? "暂无可加购 SKU"
                      : addCartMutation.isPending
                        ? "加入中..."
                        : "加入购物车"}
                </button>
              </div>
              {addCartMutation.isError ? (
                <div className="error-box">{(addCartMutation.error as Error).message}</div>
              ) : null}
              {addCartMutation.isSuccess ? (
                <div className="success-box">已加入购物车。</div>
              ) : null}
              <article className="description">
                <h2>商品描述</h2>
                <p>{product.detail || "暂无更多商品描述。"}</p>
              </article>
            </div>
          </>
        ) : null}
      </div>

      <div className="panel">
        <div className="section-head">
          <h2>用户评价</h2>
          <span className="muted">
            共 {statsQuery.data?.data?.total_count ?? 0} 条，均分{" "}
            {statsQuery.data?.data?.average_rating ?? 0}
          </span>
        </div>
        <div className="review-list">
          {(reviewsQuery.data?.data || []).map((review) => (
            <article className="review-card" key={review.id}>
              <div className="review-head">
                <strong>用户 {review.user_id}</strong>
                <span>{review.created_at || "暂无时间"}</span>
              </div>
              <div className="muted">评分：{review.rating}</div>
              <p>{review.content || "暂无文字评价"}</p>
              {review.reply_content ? <div className="reply-box">回复：{review.reply_content}</div> : null}
            </article>
          ))}
          {!reviewsQuery.isLoading && (reviewsQuery.data?.data?.length ?? 0) === 0 ? (
            <div className="muted">暂无评价</div>
          ) : null}
        </div>
        {profile ? (
          <ReviewComposer
            onSubmit={(payload) => reviewMutation.mutate(payload)}
            productId={Number(id)}
            defaultSkuId={selectedSku?.id ?? Number(id)}
            successMessage={reviewMessage}
            userId={profile.id}
          />
        ) : null}
      </div>
    </section>
  );
}

function ReviewComposer(props: {
  userId: number;
  productId: number;
  defaultSkuId: number;
  successMessage: string;
  onSubmit: (payload: {
    user_id: number;
    order_id: number;
    order_item_id: number;
    product_id: number;
    sku_id: number;
    rating: number;
    content: string;
    images?: string[];
    videos?: string[];
  }) => void;
}) {
  function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const formData = new FormData(event.currentTarget);
    props.onSubmit({
      user_id: props.userId,
      order_id: Number(formData.get("order_id") || 0),
      order_item_id: Number(formData.get("order_item_id") || 0),
      product_id: props.productId,
      sku_id: Number(formData.get("sku_id") || props.defaultSkuId),
      rating: Number(formData.get("rating") || 5),
      content: String(formData.get("content") || ""),
      images: [],
      videos: [],
    });
    event.currentTarget.reset();
  }

  return (
    <form className="form review-form" onSubmit={handleSubmit}>
      <h3>发表评价</h3>
      <p className="muted">请填写对应订单信息后提交评价。</p>
      <div className="form-grid">
        <label>
          订单 ID
          <input name="order_id" placeholder="例如 1" required />
        </label>
        <label>
          订单项 ID
          <input name="order_item_id" placeholder="例如 1" required />
        </label>
        <label>
          SKU ID
          <input name="sku_id" placeholder={`默认使用 ${props.defaultSkuId}`} />
        </label>
        <label>
          评分
          <select defaultValue="5" name="rating">
            <option value="5">5 星</option>
            <option value="4">4 星</option>
            <option value="3">3 星</option>
            <option value="2">2 星</option>
            <option value="1">1 星</option>
          </select>
        </label>
      </div>
      <label>
        评价内容
        <textarea className="text-area" name="content" placeholder="写点使用感受" required />
      </label>
      {props.successMessage ? <div className="success-box">{props.successMessage}</div> : null}
      <button className="primary-button" type="submit">
        提交评价
      </button>
    </form>
  );
}
