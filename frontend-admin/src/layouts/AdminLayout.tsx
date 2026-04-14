import { NavLink, Outlet } from "react-router-dom";
import { useAdminAuthStore } from "@/stores/adminAuth";

const menuItems = [
  { to: "/", label: "仪表盘" },
  { to: "/users", label: "用户管理" },
  { to: "/products", label: "商品管理" },
  { to: "/skus", label: "SKU 管理" },
  { to: "/categories", label: "分类管理" },
  { to: "/banners", label: "Banner 管理" },
  { to: "/orders", label: "订单管理" },
  { to: "/seckill", label: "秒杀活动" },
];

export function AdminLayout() {
  const { username, logout } = useAdminAuthStore();

  return (
    <div className="admin-shell">
      <aside className="sidebar">
        <div>
          <div className="admin-brand">Sun Weilin Admin</div>
          <div className="sidebar-hint">管理后台</div>
        </div>
        <nav className="sidebar-nav">
          {menuItems.map((item) => (
            <NavLink
              key={item.to}
              className={({ isActive }) => (isActive ? "sidebar-link active" : "sidebar-link")}
              end={item.to === "/"}
              to={item.to}
            >
              {item.label}
            </NavLink>
          ))}
        </nav>
        <button className="outline-button" onClick={logout} type="button">
          退出登录
        </button>
      </aside>
      <div className="admin-main">
        <header className="admin-header">
          <div>
            <h1>后台管理</h1>
            <p>商品、订单和活动都可以在这里管理。</p>
          </div>
          <div className="admin-user">当前用户：{username || "未命名用户"}</div>
        </header>
        <main className="admin-content">
          <Outlet />
        </main>
        <footer className="site-footer admin-footer">Made by Sun Weilin</footer>
      </div>
    </div>
  );
}
