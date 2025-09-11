import "@test/mock-auth";
import { badResponseCodes } from "@test/utils";
import { describe, expect, test } from "vitest";
import { getOrgById } from "./get-org-by-id";

describe("getOrgById should handle response status codes from GET orgs/current endpoint", () => {
  const orgId = "test-id";
  test("200 status", async () => {
    const { data: spec } = await getOrgById({ orgId });
    expect(spec).toHaveProperty("id");
    expect(spec).toHaveProperty("name");
  });

  test.each(badResponseCodes)("%s status", async (code) => {
    const { error, status } = await getOrgById({ orgId });
    expect(status).toBe(code);
    expect(error).toMatchSnapshot({
      error: expect.any(String),
      description: expect.any(String),
      user_error: expect.any(Boolean),
    });
  });
});
