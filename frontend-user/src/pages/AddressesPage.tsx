import { FormEvent, useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { addAddress, deleteAddress, getAddresses, updateAddress } from "@/api/store";
import { useAuthStore } from "@/stores/auth";
import type { Address } from "@/types";

const emptyForm: Omit<Address, "id"> = {
  user_id: 0,
  receiver_name: "",
  receiver_phone: "",
  province: "",
  city: "",
  district: "",
  detail: "",
  postal_code: "",
  is_default: 0,
};

export function AddressesPage() {
  const queryClient = useQueryClient();
  const profile = useAuthStore((state) => state.profile)!;
  const [editing, setEditing] = useState<Address | null>(null);

  const query = useQuery({
    queryKey: ["addresses", profile.id],
    queryFn: () => getAddresses(profile.id),
  });

  const saveMutation = useMutation({
    mutationFn: async (payload: Omit<Address, "id"> | Address) => {
      return "id" in payload ? updateAddress(payload) : addAddress(payload);
    },
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["addresses", profile.id] });
      setEditing(null);
    },
  });

  const deleteMutation = useMutation({
    mutationFn: (id: number) => deleteAddress(id, profile.id),
    onSuccess: () => void queryClient.invalidateQueries({ queryKey: ["addresses", profile.id] }),
  });

  function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const formData = new FormData(event.currentTarget);
    const payload = {
      ...(editing ?? {}),
      user_id: profile.id,
      receiver_name: String(formData.get("receiver_name") || ""),
      receiver_phone: String(formData.get("receiver_phone") || ""),
      province: String(formData.get("province") || ""),
      city: String(formData.get("city") || ""),
      district: String(formData.get("district") || ""),
      detail: String(formData.get("detail") || ""),
      postal_code: String(formData.get("postal_code") || ""),
      is_default: Number(formData.get("is_default") || 0),
    };
    saveMutation.mutate(payload as Address | Omit<Address, "id">);
  }

  const items = query.data?.data ?? [];

  return (
    <section className="stack two-columns">
      <div className="panel">
        <div className="section-head">
          <h1>地址管理</h1>
          <button className="ghost-button" onClick={() => setEditing(null)} type="button">
            新建地址
          </button>
        </div>
        <div className="address-list">
          {items.map((item) => (
            <article className="address-card" key={item.id}>
              <div className="review-head">
                <strong>{item.receiver_name}</strong>
                <span>{item.is_default ? "默认地址" : "普通地址"}</span>
              </div>
              <p className="muted">
                {item.receiver_phone} · {item.province} {item.city} {item.district} {item.detail}
              </p>
              <div className="hero-actions">
                <button className="ghost-button" onClick={() => setEditing(item)} type="button">
                  编辑
                </button>
                <button
                  className="ghost-button"
                  onClick={() => deleteMutation.mutate(item.id)}
                  type="button"
                >
                  删除
                </button>
              </div>
            </article>
          ))}
          {!query.isLoading && items.length === 0 ? <div className="muted">暂无地址</div> : null}
        </div>
      </div>
      <div className="panel">
        <h2>{editing ? "编辑地址" : "新增地址"}</h2>
        <form className="form" key={editing?.id ?? "new"} onSubmit={handleSubmit}>
          <label>
            收件人
            <input defaultValue={editing?.receiver_name ?? emptyForm.receiver_name} name="receiver_name" required />
          </label>
          <label>
            手机号
            <input defaultValue={editing?.receiver_phone ?? emptyForm.receiver_phone} name="receiver_phone" required />
          </label>
          <label>
            省份
            <input defaultValue={editing?.province ?? emptyForm.province} name="province" required />
          </label>
          <label>
            城市
            <input defaultValue={editing?.city ?? emptyForm.city} name="city" required />
          </label>
          <label>
            区县
            <input defaultValue={editing?.district ?? emptyForm.district} name="district" required />
          </label>
          <label>
            详细地址
            <input defaultValue={editing?.detail ?? emptyForm.detail} name="detail" required />
          </label>
          <label>
            邮编
            <input defaultValue={editing?.postal_code ?? emptyForm.postal_code} name="postal_code" />
          </label>
          <label>
            默认地址
            <select defaultValue={String(editing?.is_default ?? 0)} name="is_default">
              <option value="0">否</option>
              <option value="1">是</option>
            </select>
          </label>
          <button className="primary-button" disabled={saveMutation.isPending} type="submit">
            {saveMutation.isPending ? "保存中..." : "保存地址"}
          </button>
        </form>
      </div>
    </section>
  );
}
