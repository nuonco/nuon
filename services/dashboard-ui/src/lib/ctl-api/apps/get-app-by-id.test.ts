import "@test/mock-auth";
import { badResponseCodes } from "@test/utils";
import { describe, expect, test } from "vitest";
import { getAppById } from "./get-app-by-id";

describe("getAppById should handle response status codes from GET apps/:id endpoint", () => {
  const appId = "test-id";
  const orgId = "test-id";
  test("200 status", async () => {
    const { data: spec } = await getAppById({ appId, orgId });
    expect(spec).toHaveProperty("id");
    expect(spec).toHaveProperty("name");
  });

  test.each(badResponseCodes)("%s status", async (code) => {
    const { error, status } = await getAppById({ appId, orgId });
    expect(status).toBe(code);
    expect(error).toMatchSnapshot({
      error: expect.any(String),
      description: expect.any(String),
      user_error: expect.any(Boolean),
    });
  });
});
