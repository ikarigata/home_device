import { getIdToken } from "../auth/cognito";
import { config } from "../config";

const API = config.apiEndpoint;
const TIMEOUT_MS = 10000;

// 認証切れ（未ログイン / トークン失効による 401）を表す。呼び出し側で再ログインへ誘導する。
export class UnauthorizedError extends Error {
  constructor() {
    super("セッションの有効期限が切れました。再度ログインしてください。");
    this.name = "UnauthorizedError";
  }
}

async function authHeaders(): Promise<HeadersInit> {
  const token = await getIdToken();
  if (!token) throw new UnauthorizedError();
  return { Authorization: token };
}

async function request(path: string, init?: RequestInit): Promise<Response> {
  const controller = new AbortController();
  const timer = setTimeout(() => controller.abort(), TIMEOUT_MS);
  let res: Response;
  try {
    res = await fetch(`${API}${path}`, {
      ...init,
      headers: await authHeaders(),
      signal: controller.signal,
    });
  } catch (err) {
    if (err instanceof DOMException && err.name === "AbortError") {
      throw new Error("通信がタイムアウトしました");
    }
    throw err;
  } finally {
    clearTimeout(timer);
  }
  if (res.status === 401) throw new UnauthorizedError();
  return res;
}

// 撮影トリガを送る（ラズパイが撮影→S3 アップロードを行う）。
export async function requestCapture(): Promise<void> {
  const res = await request("/capture", { method: "POST" });
  if (!res.ok) throw new Error(`撮影リクエストに失敗しました (${res.status})`);
}

export interface ImageResponse {
  url: string;
  key: string;
  capturedAt: string; // ISO8601 UTC（S3 LastModified）
  expiresIn: number;
}

// 最新画像、または指定キーの履歴画像のメタデータと署名付き URL を取得する。
// オブジェクトがまだ存在しない場合は null を返す（撮影完了ポーリングで利用）。
export async function fetchImage(key?: string): Promise<ImageResponse | null> {
  const path = key ? `/image?key=${encodeURIComponent(key)}` : "/image";
  const res = await request(path);
  if (res.status === 404) return null;
  if (!res.ok) throw new Error(`画像 URL の取得に失敗しました (${res.status})`);
  return (await res.json()) as ImageResponse;
}

export interface HistoryItem {
  key: string;
  capturedAt: string; // ISO8601 UTC
}

// 履歴画像の一覧を新しい順に取得する。
export async function fetchImageList(limit = 50): Promise<HistoryItem[]> {
  const res = await request(`/images?limit=${limit}`);
  if (!res.ok) throw new Error(`履歴の取得に失敗しました (${res.status})`);
  const data = (await res.json()) as { items: HistoryItem[] };
  return data.items;
}
