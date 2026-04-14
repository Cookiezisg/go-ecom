import { FormEvent, useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { createCategory, deleteCategory, listCategoryTree, updateCategory } from "@/api/admin";

type CategoryForm = {
  id?: number;
  parent_id: number;
  name: string;
  level: number;
  sort: number;
  icon: string;
  image: string;
  description: string;
  status: number;
};

export function CategoriesPage() {
  const queryClient = useQueryClient();
  const [editing, setEditing] = useState<CategoryForm | null>(null);
  const query = useQuery({
    queryKey: ["admin-categories"],
    queryFn: () => listCategoryTree(1),
  });
  const saveMutation = useMutation({
    mutationFn: (payload: CategoryForm) => (payload.id ? updateCategory(payload.id, payload) : createCategory(payload)),
    onSuccess: () => {
      setEditing(null);
      void queryClient.invalidateQueries({ queryKey: ["admin-categories"] });
    },
  });
  const deleteMutation = useMutation({
    mutationFn: deleteCategory,
    onSuccess: () => void queryClient.invalidateQueries({ queryKey: ["admin-categories"] }),
  });

  function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const formData = new FormData(event.currentTarget);
    saveMutation.mutate({
      ...(editing ?? {}),
      parent_id: Number(formData.get("parent_id") || 0),
      name: String(formData.get("name") || ""),
      level: Number(formData.get("level") || 1),
      sort: Number(formData.get("sort") || 0),
      icon: String(formData.get("icon") || ""),
      image: String(formData.get("image") || ""),
      description: String(formData.get("description") || ""),
      status: Number(formData.get("status") || 1),
    });
  }

  const list = query.data?.data ?? [];

  return (
    <section className="admin-grid two-panel">
      <div className="table-card">
        <div className="card-head">
          <h2>分类管理</h2>
          <button className="outline-button" onClick={() => setEditing(null)} type="button">
            新建分类
          </button>
        </div>
        <div className="tree-list">
          {list.map((item) => (
            <div className="tree-item" key={String(item.id)}>
              <div className="tree-head">
                <strong>{String(item.name)}</strong>
                <div className="action-row">
                  <button className="table-button" onClick={() => setEditing(item as unknown as CategoryForm)} type="button">
                    编辑
                  </button>
                  <button className="table-button danger" onClick={() => deleteMutation.mutate(Number(item.id))} type="button">
                    删除
                  </button>
                </div>
              </div>
              {Array.isArray(item.children) && item.children.length > 0 ? (
                <div className="tree-children">
                  {item.children.map((child: Record<string, unknown>) => (
                    <div className="tree-child" key={String(child.id)}>
                      <span>{String(child.name)}</span>
                    </div>
                  ))}
                </div>
              ) : null}
            </div>
          ))}
        </div>
      </div>
      <div className="table-card">
        <h2>{editing ? "编辑分类" : "新建分类"}</h2>
        <form className="admin-form" key={editing?.id ?? "new-category"} onSubmit={handleSubmit}>
          <input defaultValue={editing?.name ?? ""} name="name" placeholder="分类名称" required />
          <input defaultValue={editing?.parent_id ?? 0} name="parent_id" placeholder="父分类 ID" />
          <input defaultValue={editing?.level ?? 1} name="level" placeholder="层级" />
          <input defaultValue={editing?.sort ?? 0} name="sort" placeholder="排序" />
          <input defaultValue={editing?.icon ?? ""} name="icon" placeholder="图标 URL" />
          <input defaultValue={editing?.image ?? ""} name="image" placeholder="图片 URL" />
          <textarea className="admin-text-area" defaultValue={editing?.description ?? ""} name="description" placeholder="描述" />
          <select defaultValue={String(editing?.status ?? 1)} name="status">
            <option value="0">禁用</option>
            <option value="1">启用</option>
          </select>
          <button className="primary-button" type="submit">
            保存分类
          </button>
        </form>
      </div>
    </section>
  );
}
