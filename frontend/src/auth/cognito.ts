import {
  CognitoUserPool,
  CognitoUser,
  AuthenticationDetails,
  type CognitoUserSession,
} from "amazon-cognito-identity-js";
import { config } from "../config";

let pool: CognitoUserPool | null = null;
function userPool(): CognitoUserPool {
  if (!pool) {
    pool = new CognitoUserPool({ UserPoolId: config.userPoolId, ClientId: config.clientId });
  }
  return pool;
}

export function signIn(username: string, password: string): Promise<void> {
  const user = new CognitoUser({ Username: username, Pool: userPool() });
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
  const user = userPool().getCurrentUser();
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
  userPool().getCurrentUser()?.signOut();
}
