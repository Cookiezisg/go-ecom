import { Link, NavLink, Outlet } from "react-router-dom";
import { useAuthStore } from "@/stores/auth";

const navItems = [
  { to: "/", label: "首页" },
  { to: "/products", label: "商品" },
  { to: "/seckill", label: "秒杀" },
  { to: "/cart", label: "购物车" },
  { to: "/orders", label: "订单" },
  { to: "/profile", label: "我的" },
];

export function StoreLayout() {
  const { profile, logout } = useAuthStore();

  return (
    <div className="app-shell">
      <header className="topbar">
        <Link className="brand" to="/">
          Sun Weilin Shop
        </Link>
        <nav className="nav">
          {navItems.map((item) => (
            <NavLink
              key={item.to}
              className={({ isActive }) => (isActive ? "nav-link active" : "nav-link")}
              to={item.to}
            >
              {item.label}
            </NavLink>
          ))}
        </nav>
        <div className="topbar-actions">
          {profile ? (
            <>
              <span className="muted">你好，{profile.nickname || profile.username}</span>
              <button className="ghost-button" onClick={logout} type="button">
                退出
              </button>
            </>
          ) : (
            <Link className="primary-button small" to="/login">
              登录
            </Link>
          )}
        </div>
      </header>
      <main className="page">
        <Outlet />
      </main>
      <footer className="site-footer">Made by Sun Weilin</footer>
    </div>
  );
}
