import { useEffect, useState } from "react";
import { Loader2, AlertCircle } from "lucide-react";
import { getIdToken } from "./auth/cognito";
import { missingEnv } from "./config";
import { LoginForm } from "./components/LoginForm";
import { Viewer } from "./components/Viewer";

export default function App() {
  const [authed, setAuthed] = useState<boolean | null>(null);

  useEffect(() => {
    if (missingEnv.length > 0) return;
    getIdToken().then((token) => setAuthed(token !== null));
  }, []);

  if (missingEnv.length > 0) {
    return (
      <div className="min-h-screen bg-zinc-950 flex items-center justify-center p-4 font-sans">
        <div className="max-w-md w-full bg-zinc-900/80 border border-red-500/40 rounded-2xl p-6 text-zinc-200">
          <div className="flex items-center gap-2 text-red-500 font-bold mb-3">
            <AlertCircle size={18} />
            設定エラー
          </div>
          <p className="text-sm text-zinc-400 mb-3">
            以下の環境変数が設定されていません。<code className="text-yellow-500">.env</code>{" "}
            を確認してください。
          </p>
          <ul className="text-sm font-mono text-yellow-500 list-disc list-inside">
            {missingEnv.map((k) => (
              <li key={k}>{k}</li>
            ))}
          </ul>
        </div>
      </div>
    );
  }

  if (authed === null) {
    return (
      <div className="min-h-screen bg-zinc-950 flex items-center justify-center">
        <Loader2 size={32} className="animate-spin text-yellow-500" />
      </div>
    );
  }

  return authed ? (
    <Viewer onLoggedOut={() => setAuthed(false)} />
  ) : (
    <LoginForm onLoggedIn={() => setAuthed(true)} />
  );
}
