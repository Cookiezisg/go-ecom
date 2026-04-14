import { Route, Routes } from "react-router-dom";
import { StoreLayout } from "@/components/StoreLayout";
import { HomePage } from "@/pages/HomePage";
import { LoginPage } from "@/pages/LoginPage";
import { ProductsPage } from "@/pages/ProductsPage";
import { ProductDetailPage } from "@/pages/ProductDetailPage";
import { CartPage } from "@/pages/CartPage";
import { SearchPage } from "@/pages/SearchPage";
import { CheckoutPage } from "@/pages/CheckoutPage";
import { OrdersPage } from "@/pages/OrdersPage";
import { OrderDetailPage } from "@/pages/OrderDetailPage";
import { ProfilePage } from "@/pages/ProfilePage";
import { AddressesPage } from "@/pages/AddressesPage";
import { CouponsPage } from "@/pages/CouponsPage";
import { SeckillPage } from "@/pages/SeckillPage";
import { SeckillDetailPage } from "@/pages/SeckillDetailPage";
import { AuthGuard } from "@/components/AuthGuard";

export function App() {
  return (
    <Routes>
      <Route element={<StoreLayout />}>
        <Route path="/" element={<HomePage />} />
        <Route path="/login" element={<LoginPage />} />
        <Route path="/products" element={<ProductsPage />} />
        <Route path="/products/:id" element={<ProductDetailPage />} />
        <Route path="/search" element={<SearchPage />} />
        <Route path="/seckill" element={<SeckillPage />} />
        <Route path="/seckill/:id" element={<SeckillDetailPage />} />
        <Route path="/cart" element={<CartPage />} />
        <Route
          path="/checkout"
          element={
            <AuthGuard>
              <CheckoutPage />
            </AuthGuard>
          }
        />
        <Route
          path="/orders"
          element={
            <AuthGuard>
              <OrdersPage />
            </AuthGuard>
          }
        />
        <Route
          path="/orders/:id"
          element={
            <AuthGuard>
              <OrderDetailPage />
            </AuthGuard>
          }
        />
        <Route
          path="/profile"
          element={
            <AuthGuard>
              <ProfilePage />
            </AuthGuard>
          }
        />
        <Route
          path="/addresses"
          element={
            <AuthGuard>
              <AddressesPage />
            </AuthGuard>
          }
        />
        <Route
          path="/coupons"
          element={
            <AuthGuard>
              <CouponsPage />
            </AuthGuard>
          }
        />
      </Route>
    </Routes>
  );
}
