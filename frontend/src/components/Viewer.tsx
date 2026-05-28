import { useState } from "react";
import { requestCapture, fetchImageUrl } from "../api/client";
import { signOut } from "../auth/cognito";

// 撮影トリガ後、ラズパイがアップロードするまでの待ち時間（秒）
const CAPTURE_WAIT_MS = 4000;

export function Viewer({ onLoggedOut }: { onLoggedOut: () => void }) {
  const [imageUrl, setImageUrl] = useState<string | null>(null);
  const [status, setStatus] = useState<string>("");
  const [busy, setBusy] = useState(false);

  async function reload() {
    setBusy(true);
    setStatus("最新画像を取得中...");
    try {
      const url = await fetchImageUrl();
      setImageUrl(`${url}${url.includes("?") ? "&" : "?"}t=${Date.now()}`);
      setStatus("");
    } catch (err) {
      setStatus(err instanceof Error ? err.message : "取得に失敗しました");
    } finally {
      setBusy(false);
    }
  }

  async function capture() {
    setBusy(true);
    setStatus("撮影をリクエスト中...");
    try {
      await requestCapture();
      setStatus(`撮影中... ${CAPTURE_WAIT_MS / 1000}秒後に表示します`);
      await new Promise((r) => setTimeout(r, CAPTURE_WAIT_MS));
      await reload();
    } catch (err) {
      setStatus(err instanceof Error ? err.message : "撮影に失敗しました");
      setBusy(false);
    }
  }

  function handleSignOut() {
    signOut();
    onLoggedOut();
  }

  return (
    <div style={{ display: "grid", gap: 12 }}>
      <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center" }}>
        <h1 style={{ margin: 0 }}>調味料モニター</h1>
        <button onClick={handleSignOut}>ログアウト</button>
      </div>

      <button onClick={capture} disabled={busy} style={{ padding: "16px", fontSize: 18 }}>
        撮影する
      </button>
      <button onClick={reload} disabled={busy}>
        最新画像を再読み込み
      </button>

      {status && <p>{status}</p>}

      {imageUrl ? (
        <img src={imageUrl} alt="調味料置き場" style={{ width: "100%", borderRadius: 8 }} />
      ) : (
        <p style={{ color: "#888" }}>まだ画像がありません。「撮影する」を押してください。</p>
      )}
    </div>
  );
}
