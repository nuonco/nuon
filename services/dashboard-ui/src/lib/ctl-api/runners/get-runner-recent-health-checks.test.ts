import "@test/mock-auth";
import { badResponseCodes } from "@test/utils";
import { describe, expect, test } from "vitest";
import { getRunnerRecentHealthChecks } from "./get-runner-recent-health-checks";

describe("getRunnerRecentHealthChecks should handle response status codes from GET runners/:id/recent-health-checks endpoint", () => {
  const runnerId = "test-id";
  const orgId = "test-id";
  
  test("200 status with all parameters", async () => {
    const { data: spec } = await getRunnerRecentHealthChecks({
      runnerId,
      orgId,
      limit: 10,
      offset: 0,
      window: "24h",
    });
    expect(Array.isArray(spec)).toBe(true);
  });

  test.each(badResponseCodes)("%s status", async (code) => {
    const { error, status } = await getRunnerRecentHealthChecks({ runnerId, orgId });
    expect(status).toBe(code);
    expect(error).toMatchSnapshot({
      error: expect.any(String),
      description: expect.any(String),
      user_error: expect.any(Boolean),
    });
  });
});
