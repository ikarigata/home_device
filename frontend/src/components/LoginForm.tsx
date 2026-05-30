import { useState, type FormEvent } from "react";
import { Loader2, AlertCircle, Lock } from "lucide-react";
import { signIn } from "../auth/cognito";

export function LoginForm({ onLoggedIn }: { onLoggedIn: () => void }) {
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    setError(null);
    setBusy(true);
    try {
      await signIn(username, password);
      onLoggedIn();
    } catch (err) {
      setError(err instanceof Error ? err.message : "ログインに失敗しました");
    } finally {
      setBusy(false);
    }
  }

  return (
    <div className="min-h-screen bg-zinc-950 flex flex-col items-center justify-center font-sans text-zinc-200 relative overflow-hidden p-4">
      {/* アンビエントグロウ */}
      <div className="fixed top-0 left-0 w-full h-full pointer-events-none z-0">
        <div className="absolute top-[-10%] left-[-10%] w-[60vw] h-[60vw] bg-yellow-500/10 rounded-full blur-[100px] animate-pulse"></div>
        <div
          className="absolute bottom-[-10%] right-[-10%] w-[50vw] h-[50vw] bg-yellow-600/10 rounded-full blur-[100px] animate-pulse"
          style={{ animationDelay: "2s" }}
        ></div>
      </div>

      <div className="relative z-10 w-full max-w-md bg-zinc-900/80 backdrop-blur-3xl border border-zinc-800 rounded-2xl p-6 shadow-2xl flex flex-col">
        <div className="flex flex-col items-center mb-8 mt-2">
          <div className="w-12 h-12 rounded-xl bg-yellow-500/10 border border-yellow-500/20 flex items-center justify-center mb-4">
            <Lock size={22} className="text-yellow-500" />
          </div>
          <h1 className="text-xl font-bold tracking-wide text-zinc-100">Kitchen Monitor</h1>
          <p className="text-yellow-500/80 text-xs font-bold mt-1 tracking-[0.2em] uppercase">
            Spice Stock Checker
          </p>
        </div>

        <form onSubmit={handleSubmit} className="flex flex-col gap-3">
          <input
            type="text"
            placeholder="ユーザー名"
            aria-label="ユーザー名"
            value={username}
            autoCapitalize="none"
            autoComplete="username"
            onChange={(e) => setUsername(e.target.value)}
            className="w-full px-4 py-3 rounded-xl bg-zinc-950/60 border border-zinc-700 text-zinc-100 placeholder-zinc-600 text-sm font-medium focus:outline-none focus:border-yellow-500/50 focus:shadow-[0_0_10px_rgba(234,179,8,0.15)] transition-all"
          />
          <input
            type="password"
            placeholder="パスワード"
            aria-label="パスワード"
            value={password}
            autoComplete="current-password"
            onChange={(e) => setPassword(e.target.value)}
            className="w-full px-4 py-3 rounded-xl bg-zinc-950/60 border border-zinc-700 text-zinc-100 placeholder-zinc-600 text-sm font-medium focus:outline-none focus:border-yellow-500/50 focus:shadow-[0_0_10px_rgba(234,179,8,0.15)] transition-all"
          />

          {error && (
            <div className="flex items-center gap-2 px-4 py-2.5 rounded-lg text-xs font-bold tracking-wide bg-zinc-950/50 border border-red-500/50 text-red-500 shadow-[0_0_10px_rgba(239,68,68,0.2)]">
              <AlertCircle size={14} />
              {error}
            </div>
          )}

          <button
            type="submit"
            disabled={busy}
            className="flex items-center justify-center gap-2 w-full py-4 mt-2 rounded-xl font-bold text-base tracking-wide transition-all duration-200 bg-yellow-500 hover:bg-yellow-400 active:scale-95 text-zinc-950 shadow-[0_0_20px_rgba(234,179,8,0.3)] disabled:opacity-40 disabled:cursor-not-allowed disabled:shadow-none"
          >
            {busy && <Loader2 size={18} className="animate-spin" />}
            {busy ? "ログイン中..." : "ログイン"}
          </button>
        </form>
      </div>
    </div>
  );
}
