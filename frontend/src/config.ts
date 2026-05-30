const RAW = {
  VITE_API_ENDPOINT: import.meta.env.VITE_API_ENDPOINT as string | undefined,
  VITE_USER_POOL_ID: import.meta.env.VITE_USER_POOL_ID as string | undefined,
  VITE_CLIENT_ID: import.meta.env.VITE_CLIENT_ID as string | undefined,
};

export const missingEnv = Object.entries(RAW)
  .filter(([, v]) => !v)
  .map(([k]) => k);

export const config = {
  apiEndpoint: RAW.VITE_API_ENDPOINT ?? "",
  userPoolId: RAW.VITE_USER_POOL_ID ?? "",
  clientId: RAW.VITE_CLIENT_ID ?? "",
};
