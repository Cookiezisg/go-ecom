import { FormEvent, useState } from "react";
import { useNavigate } from "react-router-dom";
import { adminLogin } from "@/api/admin";
import { useAdminAuthStore } from "@/stores/adminAuth";

export function AdminLoginPage() {
  const navigate = useNavigate();
  const setAuth = useAdminAuthStore((state) => state.setAuth);
  const [error, setError] = useState("");
  const [submitting, setSubmitting] = useState(false);

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setSubmitting(true);
    setError("");

    const formData = new FormData(event.currentTarget);

    try {
      const response = await adminLogin({
        username: String(formData.get("username") || ""),
        password: String(formData.get("password") || ""),
        login_type: 1,
      });
      setAuth(response.data.token, response.data.user.nickname || response.data.user.username);
      navigate("/");
    } catch (submitError) {
      setError(submitError instanceof Error ? submitError.message : "登录失败");
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <section className="login-screen">
      <div className="login-panel">
        <span className="chip">后台入口</span>
        <h1>登录后台</h1>
        <p>请输入账号和密码。</p>
        <form className="admin-form" onSubmit={handleSubmit}>
          <label>
            用户名
            <input name="username" placeholder="请输入用户名" required />
          </label>
          <label>
            密码
            <input name="password" placeholder="请输入密码" required type="password" />
          </label>
          {error ? <div className="error-box">{error}</div> : null}
          <button className="primary-button" disabled={submitting} type="submit">
            {submitting ? "登录中..." : "登录"}
          </button>
        </form>
      </div>
    </section>
  );
}
