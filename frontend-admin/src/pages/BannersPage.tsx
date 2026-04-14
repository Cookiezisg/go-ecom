import { FormEvent, useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { createBanner, deleteBanner, listBanners, updateBanner } from "@/api/admin";
import { uploadImage } from "@/api/upload";

type BannerForm = {
  id?: number;
  title: string;
  description: string;
  image: string;
  image_local?: string;
  link: string;
  link_type: number;
  sort: number;
  status: number;
  start_time: string;
  end_time: string;
};

export function BannersPage() {
  const queryClient = useQueryClient();
  const [editing, setEditing] = useState<BannerForm | null>(null);
  const query = useQuery({
    queryKey: ["admin-banners"],
    queryFn: () => listBanners(),
  });
  const saveMutation = useMutation({
    mutationFn: (payload: BannerForm) => (payload.id ? updateBanner(payload.id, payload) : createBanner(payload)),
    onSuccess: () => {
      setEditing(null);
      void queryClient.invalidateQueries({ queryKey: ["admin-banners"] });
    },
  });
  const deleteMutation = useMutation({
    mutationFn: deleteBanner,
    onSuccess: () => void queryClient.invalidateQueries({ queryKey: ["admin-banners"] }),
  });
  const uploadMutation = useMutation({
    mutationFn: (file: File) => uploadImage(file, "image"),
  });

  function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const formData = new FormData(event.currentTarget);
    saveMutation.mutate({
      ...(editing ?? {}),
      title: String(formData.get("title") || ""),
      description: String(formData.get("description") || ""),
      image: String(formData.get("image") || ""),
      link: String(formData.get("link") || ""),
      link_type: Number(formData.get("link_type") || 4),
      sort: Number(formData.get("sort") || 0),
      status: Number(formData.get("status") || 1),
      start_time: String(formData.get("start_time") || ""),
      end_time: String(formData.get("end_time") || ""),
    });
  }

  const list = query.data?.data ?? [];

  return (
    <section className="admin-grid two-panel">
      <div className="table-card">
        <div className="card-head">
          <h2>Banner 管理</h2>
          <button className="outline-button" onClick={() => setEditing(null)} type="button">
            新建 Banner
          </button>
        </div>
        <table className="table">
          <thead>
            <tr>
              <th>ID</th>
              <th>标题</th>
              <th>链接</th>
              <th>状态</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            {list.map((banner) => (
              <tr key={String(banner.id)}>
                <td>{String(banner.id)}</td>
                <td>{String(banner.title)}</td>
                <td>{String(banner.link || "-")}</td>
                <td>{String(banner.status)}</td>
                <td>
                  <div className="action-row">
                    <button className="table-button" onClick={() => setEditing(banner as unknown as BannerForm)} type="button">
                      编辑
                    </button>
                    <button className="table-button danger" onClick={() => deleteMutation.mutate(Number(banner.id))} type="button">
                      删除
                    </button>
                  </div>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
      <div className="table-card">
        <h2>{editing ? "编辑 Banner" : "新建 Banner"}</h2>
        <form className="admin-form" key={editing?.id ?? "new-banner"} onSubmit={handleSubmit}>
          <input defaultValue={editing?.title ?? ""} name="title" placeholder="标题" required />
          <input defaultValue={editing?.description ?? ""} name="description" placeholder="描述" />
          <input defaultValue={editing?.image ?? ""} name="image" placeholder="图片 URL" required />
          <input
            accept="image/*"
            onChange={(event) => {
              const file = event.target.files?.[0];
              if (!file) return;
              uploadMutation.mutate(file, {
                onSuccess: (response) => {
                  const target = document.querySelector<HTMLInputElement>('input[name="image"]');
                  if (target && response.data?.file_url) {
                    target.value = response.data.file_url;
                  }
                },
              });
            }}
            type="file"
          />
          <input defaultValue={editing?.link ?? ""} name="link" placeholder="跳转链接" />
          <select defaultValue={String(editing?.link_type ?? 4)} name="link_type">
            <option value="1">商品详情</option>
            <option value="2">分类页面</option>
            <option value="3">外部链接</option>
            <option value="4">无链接</option>
          </select>
          <input defaultValue={editing?.sort ?? 0} name="sort" placeholder="排序" />
          <select defaultValue={String(editing?.status ?? 1)} name="status">
            <option value="0">禁用</option>
            <option value="1">启用</option>
          </select>
          <input defaultValue={editing?.start_time ?? ""} name="start_time" placeholder="开始时间" />
          <input defaultValue={editing?.end_time ?? ""} name="end_time" placeholder="结束时间" />
          <button className="primary-button" type="submit">
            保存 Banner
          </button>
        </form>
      </div>
    </section>
  );
}
