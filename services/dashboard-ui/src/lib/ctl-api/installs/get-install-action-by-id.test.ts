import "@test/mock-auth";
import { badResponseCodes } from "@test/utils";
import { describe, expect, test } from "vitest";
import { getInstallActionById } from "./get-install-action-by-id";

describe("getInstallActionById should handle response status codes from GET installs/:installId/action-workflows/:actionId/recent-runs endpoint", () => {
  const installId = "test-install-id";
  const actionId = "test-action-id";
  const orgId = "test-org-id";

  test("200 status with pagination", async () => {
    const { data: runs, status } = await getInstallActionById({
      installId,
      actionId,
      orgId,
      limit: 10,
      offset: 0,
    });
    expect(status).toBe(200);
    expect(runs).toHaveProperty("id");
  });

  test.each(badResponseCodes)("%s status", async (code) => {
    const { error, status } = await getInstallActionById({
      installId,
      actionId,
      orgId,
    });
    expect(status).toBe(code);
    expect(error).toMatchSnapshot({
      error: expect.any(String),
      description: expect.any(String),
      user_error: expect.any(Boolean),
    });
  });
});
