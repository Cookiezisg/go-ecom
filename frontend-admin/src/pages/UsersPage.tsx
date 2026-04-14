import { useQuery } from "@tanstack/react-query";
import { listUsers } from "@/api/admin";

export function UsersPage() {
  const query = useQuery({
    queryKey: ["admin-users"],
    queryFn: () => listUsers({ page: 1, page_size: 20, status: 0 }),
  });

  const users = query.data?.data?.users ?? [];

  return (
    <section className="table-card">
      <div className="card-head">
        <h2>用户管理</h2>
        <span className="muted">已接入用户列表接口</span>
      </div>
      {query.isLoading ? <div>加载中...</div> : null}
      {query.isError ? <div className="error-box">{(query.error as Error).message}</div> : null}
      <table className="table">
        <thead>
          <tr>
            <th>ID</th>
            <th>用户名</th>
            <th>昵称</th>
            <th>联系方式</th>
            <th>状态</th>
            <th>等级</th>
          </tr>
        </thead>
        <tbody>
          {users.map((user) => (
            <tr key={user.id}>
              <td>{user.id}</td>
              <td>{user.username}</td>
              <td>{user.nickname || "-"}</td>
              <td>{user.phone || user.email || "-"}</td>
              <td>{user.status}</td>
              <td>{user.member_level ?? 0}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </section>
  );
}
