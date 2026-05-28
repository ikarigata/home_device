import { useEffect, useState } from "react";
import { getIdToken } from "./auth/cognito";
import { LoginForm } from "./components/LoginForm";
import { Viewer } from "./components/Viewer";

export default function App() {
  const [authed, setAuthed] = useState<boolean | null>(null);

  useEffect(() => {
    getIdToken().then((token) => setAuthed(token !== null));
  }, []);

  if (authed === null) {
    return <p style={{ padding: 16 }}>読み込み中...</p>;
  }

  return (
    <main style={{ maxWidth: 480, margin: "0 auto", padding: 16, fontFamily: "system-ui, sans-serif" }}>
      {authed ? (
        <Viewer onLoggedOut={() => setAuthed(false)} />
      ) : (
        <LoginForm onLoggedIn={() => setAuthed(true)} />
      )}
    </main>
  );
}
