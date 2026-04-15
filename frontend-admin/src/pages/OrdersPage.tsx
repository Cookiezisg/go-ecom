import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { listOrders } from "@/api/admin";
import { DataTableControls } from "@/components/DataTableControls";

export function OrdersPage() {
  const [keyword, setKeyword] = useState("");
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const query = useQuery({
    queryKey: ["admin-orders", page, pageSize, keyword],
    queryFn: () => listOrders({ page, page_size: pageSize, status: -1, keyword }),
  });
  const orders = query.data?.data?.list ?? [];

  return (
    <section className="table-card">
      <div className="card-head">
        <h2>订单管理</h2>
        <span className="muted">查看订单信息。</span>
      </div>
      {query.isLoading ? <div>加载中...</div> : null}
      {query.isError ? <div className="error-box">{(query.error as Error).message}</div> : null}
      <DataTableControls
        onPageChange={setPage}
        onPageSizeChange={(size) => {
          setPageSize(size);
          setPage(1);
        }}
        onSearchChange={(value) => {
          setKeyword(value);
        }}
        page={page}
        pageSize={pageSize}
        searchPlaceholder="搜索订单号、收货人、用户 ID"
        searchValue={keyword}
        total={query.data?.data?.total ?? orders.length}
      />
      <table className="table">
        <thead>
          <tr>
            <th>订单号</th>
            <th>用户</th>
            <th>状态</th>
            <th>实付金额</th>
            <th>收货人</th>
            <th>创建时间</th>
          </tr>
        </thead>
        <tbody>
          {orders.map((order) => (
            <tr key={order.id}>
              <td>{order.order_no}</td>
              <td>{order.user_id}</td>
              <td>{order.status}</td>
              <td>{order.pay_amount}</td>
              <td>{order.receiver_name || "-"}</td>
              <td>{order.created_at || "-"}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </section>
  );
}
