import { useState } from "react";
import {
  Camera,
  Loader2,
  CheckSquare,
  AlertCircle,
  Image as ImageIcon,
  LogOut,
  RefreshCw,
  History,
  Zap,
} from "lucide-react";
import {
  requestCapture,
  fetchImage,
  fetchImageList,
  UnauthorizedError,
  type HistoryItem,
  type ImageResponse,
} from "../api/client";
import { signOut } from "../auth/cognito";

// 撮影完了ポーリング: 0.5秒間隔 × 20回 = 最大10秒待つ。
const POLL_INTERVAL_MS = 500;
const POLL_MAX_ATTEMPTS = 20;
// スマホとラズパイの時計ズレ吸収用のマージン。
const CLOCK_SKEW_MS = 5000;

type Mode = "latest" | "history";
type StatusType = "idle" | "loading" | "success" | "error";
interface Status {
  type: StatusType;
  message: string;
}

export function Viewer({ onLoggedOut }: { onLoggedOut: () => void }) {
  const [mode, setMode] = useState<Mode>("latest");
  const [imageUrl, setImageUrl] = useState<string | null>(null);
  // 現在表示中の画像の撮影時刻（最新/履歴どちらのモードでも使用）。
  const [shownAt, setShownAt] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);
  const [status, setStatus] = useState<Status>({ type: "idle", message: "システム待機中" });
  const [historyItems, setHistoryItems] = useState<HistoryItem[]>([]);
  const [historyLoaded, setHistoryLoaded] = useState(false);

  function handleError(err: unknown, fallback: string) {
    if (err instanceof UnauthorizedError) {
      signOut();
      onLoggedOut();
      return;
    }
    setStatus({ type: "error", message: err instanceof Error ? err.message : fallback });
  }

  function showImage(img: ImageResponse) {
    setImageUrl(`${img.url}${img.url.includes("?") ? "&" : "?"}t=${Date.now()}`);
    setShownAt(img.capturedAt);
  }

  async function reload() {
    setBusy(true);
    setStatus({ type: "loading", message: "画像データを取得中..." });
    try {
      const img = await fetchImage();
      if (!img) {
        setStatus({ type: "idle", message: "まだ画像がありません" });
        return;
      }
      showImage(img);
      setStatus({ type: "success", message: "最新のストック状況です" });
    } catch (err) {
      handleError(err, "取得に失敗しました");
    } finally {
      setBusy(false);
    }
  }

  async function capture() {
    setBusy(true);
    setImageUrl(null);
    setShownAt(null);
    const threshold = Date.now() - CLOCK_SKEW_MS;
    try {
      setStatus({ type: "loading", message: "撮影リクエスト送信中..." });
      await requestCapture();
      setStatus({ type: "loading", message: "撮影完了を待っています..." });

      for (let i = 0; i < POLL_MAX_ATTEMPTS; i++) {
        await new Promise((r) => setTimeout(r, POLL_INTERVAL_MS));
        const img = await fetchImage();
        if (img && new Date(img.capturedAt).getTime() > threshold) {
          showImage(img);
          setStatus({ type: "success", message: "撮影完了" });
          return;
        }
      }
      setStatus({ type: "error", message: "撮影がタイムアウトしました" });
    } catch (err) {
      handleError(err, "撮影に失敗しました");
    } finally {
      setBusy(false);
    }
  }

  async function loadHistory() {
    setBusy(true);
    setStatus({ type: "loading", message: "履歴を取得中..." });
    try {
      const items = await fetchImageList(50);
      setHistoryItems(items);
      setHistoryLoaded(true);
      setStatus(
        items.length === 0
          ? { type: "idle", message: "履歴はまだありません" }
          : { type: "success", message: `${items.length}件の履歴` },
      );
    } catch (err) {
      handleError(err, "履歴の取得に失敗しました");
    } finally {
      setBusy(false);
    }
  }

  async function selectHistory(item: HistoryItem) {
    setBusy(true);
    setImageUrl(null);
    setStatus({ type: "loading", message: "画像を取得中..." });
    try {
      const img = await fetchImage(item.key);
      if (!img) {
        setStatus({ type: "error", message: "画像が見つかりませんでした" });
        return;
      }
      showImage(img);
      setStatus({ type: "success", message: "履歴画像を表示中" });
    } catch (err) {
      handleError(err, "画像の取得に失敗しました");
    } finally {
      setBusy(false);
    }
  }

  function switchMode(next: Mode) {
    if (next === mode) return;
    setMode(next);
    setImageUrl(null);
    setShownAt(null);
    setStatus({ type: "idle", message: next === "latest" ? "システム待機中" : "履歴モード" });
    if (next === "history" && !historyLoaded) {
      void loadHistory();
    }
  }

  function handleSignOut() {
    signOut();
    onLoggedOut();
  }

  return (
    <div className="min-h-screen bg-zinc-950 flex flex-col items-center justify-center font-sans text-zinc-200 relative overflow-hidden p-4">
      <div className="fixed top-0 left-0 w-full h-full pointer-events-none z-0">
        <div className="absolute top-[-10%] left-[-10%] w-[60vw] h-[60vw] bg-yellow-500/10 rounded-full blur-[100px] animate-pulse"></div>
        <div
          className="absolute bottom-[-10%] right-[-10%] w-[50vw] h-[50vw] bg-yellow-600/10 rounded-full blur-[100px] animate-pulse"
          style={{ animationDelay: "2s" }}
        ></div>
      </div>

      <div className="relative z-10 w-full max-w-md bg-zinc-900/80 backdrop-blur-3xl border border-zinc-800 rounded-2xl p-6 shadow-2xl flex flex-col">
        <div className="flex items-center justify-between mb-5 mt-2">
          <div>
            <h1 className="text-xl font-bold tracking-wide text-zinc-100">Kitchen Monitor</h1>
            <p className="text-yellow-500/80 text-xs font-bold mt-0.5 tracking-[0.2em] uppercase">
              Spice Stock Checker
            </p>
          </div>
          <button
            onClick={handleSignOut}
            className="flex items-center gap-1.5 text-zinc-500 hover:text-zinc-300 text-xs font-medium transition-colors"
          >
            <LogOut size={14} />
            ログアウト
          </button>
        </div>

        {/* モード切り替えタブ */}
        <div className="flex gap-1 mb-5 p-1 rounded-xl bg-zinc-950/60 border border-zinc-800">
          <ModeTab
            active={mode === "latest"}
            onClick={() => switchMode("latest")}
            icon={<Zap size={14} />}
            label="最新"
          />
          <ModeTab
            active={mode === "history"}
            onClick={() => switchMode("history")}
            icon={<History size={14} />}
            label="履歴"
          />
        </div>

        {/* 画像表示エリア */}
        <div className="relative w-full aspect-video bg-zinc-950 rounded-xl overflow-hidden mb-5 flex flex-col items-center justify-center border border-zinc-800 shadow-inner">
          <div
            className={`absolute inset-0 border-2 rounded-xl pointer-events-none transition-colors duration-500 z-10 ${
              imageUrl ? "border-yellow-500/30" : "border-transparent"
            }`}
          ></div>

          {imageUrl ? (
            <img
              src={imageUrl}
              alt="キッチンの調味料"
              className="w-full h-full object-cover opacity-90 transition-opacity duration-700"
              onError={() => {
                setImageUrl(null);
                setStatus({ type: "error", message: "画像の表示に失敗しました" });
              }}
            />
          ) : (
            <div className="text-zinc-600 flex flex-col items-center">
              {busy ? (
                <Loader2 size={40} className="animate-spin text-yellow-500 mb-3" />
              ) : (
                <ImageIcon size={40} className="mb-3 opacity-30 text-zinc-500" strokeWidth={1.5} />
              )}
              <span className="text-xs font-medium tracking-widest text-zinc-500">
                {busy ? "SYNCING..." : mode === "history" ? "SELECT ITEM" : "NO IMAGE"}
              </span>
            </div>
          )}

          {imageUrl && shownAt && (
            <div className="absolute bottom-3 right-3 bg-zinc-950/80 backdrop-blur-md text-yellow-400 text-[10px] font-bold tracking-wider px-2.5 py-1.5 rounded-md border border-yellow-500/20 z-20">
              {formatTimestamp(shownAt)}
            </div>
          )}
        </div>

        {/* ステータス */}
        <div className="flex justify-center mb-5">
          <div
            className={`flex items-center gap-2 px-4 py-2.5 rounded-lg text-xs font-bold tracking-wide transition-all duration-300 bg-zinc-950/50 border ${
              status.type === "loading"
                ? "border-yellow-500/50 text-yellow-500 shadow-[0_0_10px_rgba(234,179,8,0.2)]"
                : status.type === "success"
                  ? "border-zinc-700 text-zinc-300"
                  : status.type === "error"
                    ? "border-red-500/50 text-red-500 shadow-[0_0_10px_rgba(239,68,68,0.2)]"
                    : "border-zinc-800 text-zinc-500"
            }`}
          >
            {status.type === "loading" && <Loader2 size={14} className="animate-spin" />}
            {status.type === "success" && <CheckSquare size={14} className="text-green-500" />}
            {status.type === "error" && <AlertCircle size={14} />}
            {status.message}
          </div>
        </div>

        {/* モード別コントロール */}
        {mode === "latest" ? (
          <div className="flex flex-col gap-3">
            <button
              onClick={capture}
              disabled={busy}
              className="flex items-center justify-center gap-3 w-full py-4 rounded-xl font-bold text-base tracking-wide transition-all duration-200 bg-yellow-500 hover:bg-yellow-400 active:scale-95 text-zinc-950 shadow-[0_0_20px_rgba(234,179,8,0.3)] disabled:opacity-40 disabled:cursor-not-allowed disabled:shadow-none"
            >
              {busy ? <Loader2 size={20} className="animate-spin" /> : <Camera size={20} />}
              撮影する
            </button>

            <button
              onClick={reload}
              disabled={busy}
              className="flex items-center justify-center gap-2 w-full py-3 rounded-xl font-semibold text-sm tracking-wide transition-all duration-200 border border-zinc-700 text-zinc-400 hover:text-zinc-200 hover:border-zinc-500 active:scale-95 disabled:opacity-40 disabled:cursor-not-allowed"
            >
              <RefreshCw size={15} />
              最新画像を再読み込み
            </button>
          </div>
        ) : (
          <HistoryList
            items={historyItems}
            loaded={historyLoaded}
            busy={busy}
            selectedAt={shownAt}
            onSelect={selectHistory}
            onReload={loadHistory}
          />
        )}
      </div>
    </div>
  );
}

function ModeTab({
  active,
  onClick,
  icon,
  label,
}: {
  active: boolean;
  onClick: () => void;
  icon: React.ReactNode;
  label: string;
}) {
  return (
    <button
      onClick={onClick}
      className={`flex-1 flex items-center justify-center gap-1.5 py-2 rounded-lg text-xs font-bold tracking-wide transition-all ${
        active
          ? "bg-yellow-500/15 text-yellow-400 border border-yellow-500/30"
          : "text-zinc-500 hover:text-zinc-300 border border-transparent"
      }`}
    >
      {icon}
      {label}
    </button>
  );
}

function HistoryList({
  items,
  loaded,
  busy,
  selectedAt,
  onSelect,
  onReload,
}: {
  items: HistoryItem[];
  loaded: boolean;
  busy: boolean;
  selectedAt: string | null;
  onSelect: (item: HistoryItem) => void;
  onReload: () => void;
}) {
  return (
    <div className="flex flex-col gap-3">
      <div className="max-h-64 overflow-y-auto rounded-xl border border-zinc-800 bg-zinc-950/50 divide-y divide-zinc-800/60">
        {!loaded ? (
          <div className="p-4 text-center text-xs text-zinc-500">読み込み中...</div>
        ) : items.length === 0 ? (
          <div className="p-4 text-center text-xs text-zinc-500">まだ履歴がありません</div>
        ) : (
          items.map((item) => {
            const isSelected = selectedAt === item.capturedAt;
            return (
              <button
                key={item.key}
                onClick={() => onSelect(item)}
                disabled={busy}
                className={`w-full flex items-center justify-between px-4 py-3 text-left text-sm transition-colors disabled:opacity-50 ${
                  isSelected
                    ? "bg-yellow-500/10 text-yellow-300"
                    : "text-zinc-300 hover:bg-zinc-800/60"
                }`}
              >
                <span className="font-medium">{formatTimestamp(item.capturedAt)}</span>
                {isSelected && <CheckSquare size={14} className="text-yellow-400" />}
              </button>
            );
          })
        )}
      </div>

      <button
        onClick={onReload}
        disabled={busy}
        className="flex items-center justify-center gap-2 w-full py-3 rounded-xl font-semibold text-sm tracking-wide transition-all duration-200 border border-zinc-700 text-zinc-400 hover:text-zinc-200 hover:border-zinc-500 active:scale-95 disabled:opacity-40 disabled:cursor-not-allowed"
      >
        <RefreshCw size={15} />
        履歴を再読み込み
      </button>
    </div>
  );
}

function formatTimestamp(iso: string): string {
  const d = new Date(iso);
  if (isNaN(d.getTime())) return iso;
  return d.toLocaleString([], {
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
  });
}
