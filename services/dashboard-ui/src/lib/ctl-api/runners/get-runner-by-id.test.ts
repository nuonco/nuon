import "@test/mock-auth";
import { badResponseCodes } from "@test/utils";
import { describe, expect, test } from "vitest";
import { getRunnerById } from "./get-runner-by-id";

describe("getRunnerById should handle response status codes from GET runners/:id endpoint", () => {
  const runnerId = "test-id";
  const orgId = "test-id";
  test("200 status", async () => {
    const { data: runner } = await getRunnerById({ runnerId, orgId });
    expect(runner).toHaveProperty("id");
    expect(runner).toHaveProperty("created_at");
  });

  test.each(badResponseCodes)("%s status", async (code) => {
    const { error, status } = await getRunnerById({ runnerId, orgId });
    expect(status).toBe(code);
    expect(error).toMatchSnapshot({
      error: expect.any(String),
      description: expect.any(String),
      user_error: expect.any(Boolean),
    });
  });
});
