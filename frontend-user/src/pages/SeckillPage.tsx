import { useQuery } from "@tanstack/react-query";
import { Link } from "react-router-dom";
import { listSeckillActivities } from "@/api/store";

export function SeckillPage() {
  const query = useQuery({
    queryKey: ["seckill-list"],
    queryFn: () => listSeckillActivities(1),
  });

  const list = query.data?.data?.list ?? [];

  return (
    <section className="stack">
      <div className="section-head">
        <h1>秒杀专区</h1>
        <span className="muted">限时好价，先到先得。</span>
      </div>
      <div className="product-grid">
        {list.map((item) => (
          <Link className="product-card" key={item.id} to={`/seckill/${item.id}`}>
            <div className="product-image">
              {item.sku_image ? <img alt={item.name} src={item.sku_image} /> : <span>SECKILL</span>}
            </div>
            <div className="product-content">
              <h3>{item.name}</h3>
              <p>{item.sku_name}</p>
              <div className="price-row">
                <strong>¥{item.seckill_price}</strong>
                <span>¥{item.original_price}</span>
              </div>
            </div>
          </Link>
        ))}
      </div>
    </section>
  );
}
