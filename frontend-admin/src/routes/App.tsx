import { Navigate, Route, Routes } from "react-router-dom";
import { AdminLayout } from "@/layouts/AdminLayout";
import { AdminLoginPage } from "@/pages/AdminLoginPage";
import { DashboardPage } from "@/pages/DashboardPage";
import { UsersPage } from "@/pages/UsersPage";
import { ProductsAdminPage } from "@/pages/ProductsAdminPage";
import { OrdersPage } from "@/pages/OrdersPage";
import { SeckillPage } from "@/pages/SeckillPage";
import { SkusPage } from "@/pages/SkusPage";
import { CategoriesPage } from "@/pages/CategoriesPage";
import { BannersPage } from "@/pages/BannersPage";
import { useAdminAuthStore } from "@/stores/adminAuth";

function Guard({ children }: { children: JSX.Element }) {
  const token = useAdminAuthStore((state) => state.token);
  if (!token) {
    return <Navigate replace to="/login" />;
  }
  return children;
}

export function App() {
  return (
    <Routes>
      <Route path="/login" element={<AdminLoginPage />} />
      <Route
        path="/"
        element={
          <Guard>
            <AdminLayout />
          </Guard>
        }
      >
        <Route index element={<DashboardPage />} />
        <Route path="users" element={<UsersPage />} />
        <Route path="products" element={<ProductsAdminPage />} />
        <Route path="skus" element={<SkusPage />} />
        <Route path="categories" element={<CategoriesPage />} />
        <Route path="banners" element={<BannersPage />} />
        <Route path="orders" element={<OrdersPage />} />
        <Route path="seckill" element={<SeckillPage />} />
      </Route>
    </Routes>
  );
}
