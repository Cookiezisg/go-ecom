import { FormEvent, useState } from "react";
import { useNavigate } from "react-router-dom";
import { login, register } from "@/api/store";
import { useAuthStore } from "@/stores/auth";

export function LoginPage() {
  const navigate = useNavigate();
  const setAuth = useAuthStore((state) => state.setAuth);
  const [mode, setMode] = useState<"login" | "register">("login");
  const [error, setError] = useState("");
  const [submitting, setSubmitting] = useState(false);

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setSubmitting(true);
    setError("");

    const formData = new FormData(event.currentTarget);

    try {
      if (mode === "login") {
        const response = await login({
          username: String(formData.get("username") || ""),
          password: String(formData.get("password") || ""),
          login_type: Number(formData.get("login_type") || 1),
        });

        setAuth(response.data.token, response.data.user);
        navigate("/");
      } else {
        await register({
          username: String(formData.get("username") || ""),
          password: String(formData.get("password") || ""),
          phone: String(formData.get("phone") || ""),
          email: String(formData.get("email") || ""),
        });
        setMode("login");
      }
    } catch (submitError) {
      setError(submitError instanceof Error ? submitError.message : "提交失败");
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <section className="auth-shell">
      <div className="auth-card">
        <div className="auth-switch">
          <button
            className={mode === "login" ? "tab-button active" : "tab-button"}
            onClick={() => setMode("login")}
            type="button"
          >
            登录
          </button>
          <button
            className={mode === "register" ? "tab-button active" : "tab-button"}
            onClick={() => setMode("register")}
            type="button"
          >
            注册
          </button>
        </div>
        <h1>{mode === "login" ? "登录商城" : "创建账号"}</h1>
        <p className="muted">{mode === "login" ? "欢迎回来。" : "注册后即可开始购物。"}</p>
        <form className="form" onSubmit={handleSubmit}>
          <label>
            用户名 / 手机 / 邮箱
            <input name="username" placeholder="请输入账号" required />
          </label>
          <label>
            密码
            <input name="password" placeholder="请输入密码" required type="password" />
          </label>
          {mode === "login" ? (
            <label>
              登录类型
              <select defaultValue="1" name="login_type">
                <option value="1">用户名</option>
                <option value="2">手机号</option>
                <option value="3">邮箱</option>
              </select>
            </label>
          ) : (
            <>
              <label>
                手机号
                <input name="phone" placeholder="选填但建议填写" />
              </label>
              <label>
                邮箱
                <input name="email" placeholder="选填但建议填写" type="email" />
              </label>
            </>
          )}
          {error ? <div className="error-box">{error}</div> : null}
          <button className="primary-button" disabled={submitting} type="submit">
            {submitting ? "提交中..." : mode === "login" ? "登录" : "注册"}
          </button>
        </form>
      </div>
    </section>
  );
}
