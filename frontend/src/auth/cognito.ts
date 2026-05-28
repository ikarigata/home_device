import {
  CognitoUserPool,
  CognitoUser,
  AuthenticationDetails,
  type CognitoUserSession,
} from "amazon-cognito-identity-js";

const userPool = new CognitoUserPool({
  UserPoolId: import.meta.env.VITE_USER_POOL_ID,
  ClientId: import.meta.env.VITE_CLIENT_ID,
});

export function signIn(username: string, password: string): Promise<void> {
  const user = new CognitoUser({ Username: username, Pool: userPool });
  const details = new AuthenticationDetails({ Username: username, Password: password });
  return new Promise((resolve, reject) => {
    user.authenticateUser(details, {
      onSuccess: () => resolve(),
      onFailure: (err) => reject(err),
      newPasswordRequired: () =>
        reject(new Error("初回パスワード変更が必要です。管理コンソールでパスワードを確定してください。")),
    });
  });
}

// API Gateway の Cognito JWT オーソライザーは aud(=ClientId) を持つ ID トークンを検証する。
export function getIdToken(): Promise<string | null> {
  const user = userPool.getCurrentUser();
  if (!user) return Promise.resolve(null);
  return new Promise((resolve) => {
    user.getSession((err: Error | null, session: CognitoUserSession | null) => {
      if (err || !session || !session.isValid()) {
        resolve(null);
        return;
      }
      resolve(session.getIdToken().getJwtToken());
    });
  });
}

export function signOut(): void {
  userPool.getCurrentUser()?.signOut();
}
