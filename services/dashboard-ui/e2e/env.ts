function required(name: string): string {
  const value = process.env[name];
  if (!value) throw new Error(`Missing required env var: ${name}`);
  return value;
}

export const env = {
  baseUrl: process.env.E2E_BASE_URL ?? "http://localhost:4000",
  adminApiUrl: process.env.E2E_ADMIN_API_URL ?? "http://localhost:8082",
  get adminEmail() {
    return required("E2E_ADMIN_EMAIL");
  },
  get userEmail() {
    return required("E2E_USER_EMAIL");
  },
  get orgId() {
    return required("E2E_ORG_ID");
  },
};
