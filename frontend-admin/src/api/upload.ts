import { apiClient } from "@/api/client";

export async function uploadImage(file: File, category = "image") {
  const formData = new FormData();
  formData.append("file", file);
  formData.append("category", category);
  const response = await apiClient.post<{
    code: number;
    message: string;
    data?: {
      file_id?: string;
      file_name?: string;
      file_url?: string;
    };
  }>("/api/v1/files/upload", formData, {
    headers: {
      "Content-Type": "multipart/form-data",
    },
    timeout: 60000,
  });
  const payload = response.data as unknown as Record<string, unknown>;
  const data = (payload.data as Record<string, unknown> | undefined) || {};
  return {
    ...(payload as object),
    data: {
      file_id: String(data.file_id ?? data.fileId ?? ""),
      file_name: String(data.file_name ?? data.fileName ?? ""),
      file_url: String(data.file_url ?? data.fileUrl ?? ""),
    },
  };
}
