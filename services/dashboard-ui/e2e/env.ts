function required(name: string): string {
  const value = process.env[name];
  if (!value) throw new Error(`Missing required env var: ${name}`);
  return value;
}

export const env = {
  baseUrl: process.env.E2E_BASE_URL ?? "http://127.0.0.1:4000",
  adminApiUrl: process.env.E2E_ADMIN_API_URL ?? "http://127.0.0.1:8082",
  publicApiUrl: process.env.E2E_PUBLIC_API_URL ?? "http://127.0.0.1:8081",
  get email() {
    return required("E2E_EMAIL");
  },
  orgId: process.env.E2E_ORG_ID,
  appConfig: process.env.E2E_APP_CONFIG ?? "httpbin",
};
