import { getIdToken } from "../auth/cognito";

const API = import.meta.env.VITE_API_ENDPOINT;

async function authHeaders(): Promise<HeadersInit> {
  const token = await getIdToken();
  if (!token) throw new Error("未ログインです");
  return { Authorization: token };
}

// 撮影トリガを送る（ラズパイが撮影→S3 アップロードを行う）。
export async function requestCapture(): Promise<void> {
  const res = await fetch(`${API}/capture`, {
    method: "POST",
    headers: await authHeaders(),
  });
  if (!res.ok) throw new Error(`撮影リクエストに失敗しました (${res.status})`);
}

// 最新画像の署名付き URL を取得する。
export async function fetchImageUrl(): Promise<string> {
  const res = await fetch(`${API}/image`, { headers: await authHeaders() });
  if (!res.ok) throw new Error(`画像 URL の取得に失敗しました (${res.status})`);
  const data = (await res.json()) as { url: string };
  return data.url;
}
