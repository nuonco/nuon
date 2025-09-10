import { afterAll, vi } from "vitest";

vi.mock("../src/lib/auth", async (og) => {
  const mod = await og<typeof import("../src/lib/auth")>();
  return {
    ...mod,
    auth0: {
      getSession: vi.fn().mockResolvedValue({
        tokenSet: {
          accessToken: "0000000000",
        },
      }),
    },
  };
});

afterAll(() => {
  vi.restoreAllMocks();
});
