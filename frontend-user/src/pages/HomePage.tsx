import { useQuery } from "@tanstack/react-query";
import { Link } from "react-router-dom";
import { listBanners, listProducts, listSeckillActivities } from "@/api/store";

export function HomePage() {
  const banners = useQuery({
    queryKey: ["home-banners"],
    queryFn: listBanners,
  });
  const hotProducts = useQuery({
    queryKey: ["home-hot-products"],
    queryFn: () => listProducts({ page: 1, page_size: 8, is_hot: 1, status: 1 }),
  });

  const latestProducts = useQuery({
    queryKey: ["home-latest-products"],
    queryFn: () => listProducts({ page: 1, page_size: 8, sort: "created_desc", status: 1 }),
  });
  const seckill = useQuery({
    queryKey: ["home-seckill"],
    queryFn: () => listSeckillActivities(1),
  });

  return (
    <div className="stack">
      <section className="hero">
        <div>
          <span className="eyebrow">精选好物</span>
          <h1>发现你想买的东西。</h1>
          <p>新品、热卖和限时活动都在这里。</p>
          <div className="hero-actions">
            <Link className="primary-button" to="/products">
              立即选购
            </Link>
            <Link className="secondary-button" to="/login">
              登录
            </Link>
          </div>
        </div>
        <div className="hero-panel">
          <div className="metric-card">
            <strong>Hot</strong>
            <span>热门推荐</span>
          </div>
          <div className="metric-card">
            <strong>New</strong>
            <span>最新上架</span>
          </div>
          <div className="metric-card">
            <strong>Sale</strong>
            <span>限时秒杀</span>
          </div>
        </div>
      </section>

      <section className="section">
        <div className="section-head">
          <h2>轮播推荐</h2>
          <Link to="/products">更多活动</Link>
        </div>
        <div className="banner-strip">
          {(banners.data?.data ?? []).map((banner) => (
            <article className="banner-card" key={banner.id}>
              <strong>{banner.title}</strong>
              <p>{banner.description || "活动 Banner"}</p>
            </article>
          ))}
          {!banners.isLoading && (banners.data?.data?.length ?? 0) === 0 ? (
            <div className="panel">暂无 Banner 数据</div>
          ) : null}
        </div>
      </section>

      <section className="section">
        <div className="section-head">
          <h2>热门商品</h2>
          <Link to="/products">查看全部</Link>
        </div>
        <ProductGrid
          items={hotProducts.data?.data?.list ?? []}
          loading={hotProducts.isLoading}
          emptyText="当前没有热门商品数据"
        />
      </section>

      <section className="section">
        <div className="section-head">
          <h2>最新上架</h2>
          <Link to="/products?sort=created_desc">按最新排序</Link>
        </div>
        <ProductGrid
          items={latestProducts.data?.data?.list ?? []}
          loading={latestProducts.isLoading}
          emptyText="当前没有最新商品数据"
        />
      </section>

      <section className="section">
        <div className="section-head">
          <h2>秒杀活动</h2>
          <Link to="/seckill">进入秒杀专区</Link>
        </div>
        <ProductGrid
          items={(seckill.data?.data?.list ?? []).map((item) => ({
            id: item.id,
            name: item.name,
            subtitle: item.sku_name,
            price: Number(item.seckill_price),
            original_price: Number(item.original_price),
            main_image: item.sku_image,
          }))}
          loading={seckill.isLoading}
          emptyText="当前没有秒杀活动"
          pathPrefix="/seckill"
        />
      </section>
    </div>
  );
}

function ProductGrid(props: {
  items: Array<{
    id: number;
    name: string;
    subtitle?: string;
    price: number;
    original_price?: number;
    main_image?: string;
    local_main_image?: string;
  }>;
  loading: boolean;
  emptyText: string;
  pathPrefix?: string;
}) {
  if (props.loading) {
    return <div className="panel">正在加载商品数据...</div>;
  }

  if (props.items.length === 0) {
    return <div className="panel">{props.emptyText}</div>;
  }

  return (
    <div className="product-grid">
      {props.items.map((item) => (
        <Link className="product-card" key={item.id} to={`${props.pathPrefix || "/products"}/${item.id}`}>
          <div className="product-image">
            {item.main_image || item.local_main_image ? (
              <img alt={item.name} src={item.main_image || item.local_main_image} />
            ) : (
              <span>NO IMAGE</span>
            )}
          </div>
          <div className="product-content">
            <h3>{item.name}</h3>
            <p>{item.subtitle || "品质商品，安心选购"}</p>
            <div className="price-row">
              <strong>¥{item.price}</strong>
              {item.original_price ? <span>¥{item.original_price}</span> : null}
            </div>
          </div>
        </Link>
      ))}
    </div>
  );
}
