import { useEffect, useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { Link, useSearchParams } from "react-router-dom";
import { getHotKeywords, getSearchSuggestions, searchProducts } from "@/api/store";

export function SearchPage() {
  const [searchParams, setSearchParams] = useSearchParams();
  const [keyword, setKeyword] = useState(searchParams.get("keyword") || "");
  const activeKeyword = searchParams.get("keyword") || "";

  const resultQuery = useQuery({
    queryKey: ["search-products", activeKeyword],
    queryFn: () => searchProducts({ keyword: activeKeyword, page: 1, page_size: 20, sort_by: "score_desc" }),
    enabled: Boolean(activeKeyword),
  });

  const suggestionsQuery = useQuery({
    queryKey: ["search-suggestions", keyword],
    queryFn: () => getSearchSuggestions(keyword, 6),
    enabled: keyword.trim().length > 0,
  });

  const hotKeywordsQuery = useQuery({
    queryKey: ["hot-keywords"],
    queryFn: () => getHotKeywords(8),
  });

  useEffect(() => {
    setKeyword(searchParams.get("keyword") || "");
  }, [searchParams]);

  return (
    <section className="stack">
      <div className="panel">
        <form
          className="toolbar-form"
          onSubmit={(event) => {
            event.preventDefault();
            setSearchParams(keyword ? { keyword } : {});
          }}
        >
          <input onChange={(e) => setKeyword(e.target.value)} placeholder="输入关键词搜索商品" value={keyword} />
          <div className="muted">搜索商品</div>
          <button className="primary-button" type="submit">
            搜索
          </button>
        </form>
      </div>

      {keyword && (suggestionsQuery.data?.data?.length ?? 0) > 0 ? (
        <div className="panel">
          <h2>搜索建议</h2>
          <div className="filter-chips">
            {suggestionsQuery.data?.data.map((item: string) => (
              <button className="tab-button" key={item} onClick={() => setSearchParams({ keyword: item })} type="button">
                {item}
              </button>
            ))}
          </div>
        </div>
      ) : null}

      <div className="panel">
        <h2>热词</h2>
        <div className="filter-chips">
          {(hotKeywordsQuery.data?.data ?? []).map((item: string) => (
            <button className="tab-button" key={item} onClick={() => setSearchParams({ keyword: item })} type="button">
              {item}
            </button>
          ))}
        </div>
      </div>

      <div className="product-grid">
        {(resultQuery.data?.data ?? []).map((item) => (
          <Link className="product-card" key={item.product_id} to={`/products/${item.product_id}`}>
            <div className="product-image">
              {item.main_image ? <img alt={item.name} src={item.main_image} /> : <span>SEARCH</span>}
            </div>
            <div className="product-content">
              <h3>{item.name}</h3>
              <p>相关度 {item.score.toFixed(2)}</p>
              <div className="price-row">
                <strong>¥{item.price}</strong>
                <span>销量 {item.sales}</span>
              </div>
            </div>
          </Link>
        ))}
      </div>

      {activeKeyword && !resultQuery.isLoading && (resultQuery.data?.data?.length ?? 0) === 0 ? (
        <div className="panel">没有搜索到相关商品</div>
      ) : null}
    </section>
  );
}
