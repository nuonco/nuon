import "@test/mock-auth";
import { badResponseCodes } from "@test/utils";
import { describe, expect, test } from "vitest";
import { getRunnerJobs } from "./get-runner-jobs";

describe("getRunnerJobs should handle response status codes from GET runners/:id/jobs endpoint", () => {
  const runnerId = "test-id";
  const orgId = "test-id";
  test("200 status with pagination", async () => {
    const { data: spec } = await getRunnerJobs({
      runnerId,
      orgId,
      limit: 10,
      offset: 0,
    });
    expect(Array.isArray(spec)).toBe(true);
  });

  test.each(badResponseCodes)("%s status", async (code) => {
    const { error, status } = await getRunnerJobs({ runnerId, orgId });
    expect(status).toBe(code);
    expect(error).toMatchSnapshot({
      error: expect.any(String),
      description: expect.any(String),
      user_error: expect.any(Boolean),
    });
  });
});
