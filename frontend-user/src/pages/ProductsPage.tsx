import { useMemo, useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { Link, useSearchParams } from "react-router-dom";
import { listProducts } from "@/api/store";

const sortOptions = [
  { label: "最新上架", value: "created_desc" },
  { label: "销量最高", value: "sales_desc" },
  { label: "价格升序", value: "price_asc" },
  { label: "价格降序", value: "price_desc" },
];

export function ProductsPage() {
  const [searchParams, setSearchParams] = useSearchParams();
  const [keyword, setKeyword] = useState(searchParams.get("keyword") || "");
  const page = Number(searchParams.get("page") || 1);
  const sort = searchParams.get("sort") || "created_desc";

  const query = useQuery({
    queryKey: ["products", page, sort, searchParams.get("keyword") || ""],
    queryFn: () =>
      listProducts({
        page,
        page_size: 12,
        status: 1,
        sort,
        keyword: searchParams.get("keyword") || undefined,
      }),
  });

  const list = useMemo(() => query.data?.data?.list ?? [], [query.data]);
  const totalPages = query.data?.data?.total_pages ?? 1;

  return (
    <section className="stack">
      <div className="toolbar panel">
        <form
          className="toolbar-form"
          onSubmit={(event) => {
            event.preventDefault();
            setSearchParams({ keyword, sort, page: "1" });
          }}
        >
          <input
            onChange={(event) => setKeyword(event.target.value)}
            placeholder="搜索商品名称"
            value={keyword}
          />
          <select
            onChange={(event) =>
              setSearchParams({
                keyword: searchParams.get("keyword") || "",
                sort: event.target.value,
                page: "1",
              })
            }
            value={sort}
          >
            {sortOptions.map((option) => (
              <option key={option.value} value={option.value}>
                {option.label}
              </option>
            ))}
          </select>
          <button className="primary-button" type="submit">
            搜索
          </button>
        </form>
      </div>

      {query.isLoading ? <div className="panel">正在加载商品列表...</div> : null}
      {query.isError ? <div className="error-box">{(query.error as Error).message}</div> : null}

      <div className="product-grid">
        {list.map((item) => (
          <Link className="product-card" key={item.id} to={`/products/${item.id}`}>
            <div className="product-image">
              {item.main_image || item.local_main_image ? (
                <img alt={item.name} src={item.main_image || item.local_main_image} />
              ) : (
                <span>NO IMAGE</span>
              )}
            </div>
          <div className="product-content">
            <h3>{item.name}</h3>
            <p>{item.subtitle || "优选商品，欢迎选购"}</p>
            <div className="price-row">
              <strong>¥{item.price}</strong>
              <span>销量 {item.sales}</span>
              </div>
            </div>
          </Link>
        ))}
      </div>

      <div className="pagination panel">
        <button
          className="ghost-button"
          disabled={page <= 1}
          onClick={() =>
            setSearchParams({
              keyword: searchParams.get("keyword") || "",
              sort,
              page: String(page - 1),
            })
          }
          type="button"
        >
          上一页
        </button>
        <span className="muted">
          第 {page} / {totalPages} 页
        </span>
        <button
          className="ghost-button"
          disabled={page >= totalPages}
          onClick={() =>
            setSearchParams({
              keyword: searchParams.get("keyword") || "",
              sort,
              page: String(page + 1),
            })
          }
          type="button"
        >
          下一页
        </button>
      </div>
    </section>
  );
}
