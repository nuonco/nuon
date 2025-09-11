import "@test/mock-auth";
import { badResponseCodes } from "@test/utils";
import { describe, expect, test } from "vitest";
import { getInstallActionRunById } from "./get-install-action-run-by-id";

describe("getInstallActionRunById should handle response status codes from GET installs/:id/action-workflows/runs/:runId endpoint", () => {
  const installId = "test-install-id";
  const runId = "test-run-id";
  const orgId = "test-org-id";

  test("200 status", async () => {
    const { data: run, status } = await getInstallActionRunById({
      installId,
      runId,
      orgId,
    });
    expect(status).toBe(200);
    expect(run).toHaveProperty("id");
  });

  test.each(badResponseCodes)("%s status", async (code) => {
    const { error, status } = await getInstallActionRunById({
      installId,
      runId,
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
