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
  return response.data;
}
