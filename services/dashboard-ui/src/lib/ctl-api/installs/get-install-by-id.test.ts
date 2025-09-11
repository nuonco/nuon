import "@test/mock-auth";
import { badResponseCodes } from "@test/utils";
import { describe, expect, test } from "vitest";
import { getInstallById } from "./get-install-by-id";

describe("getInstallById should handle response status codes from GET installs/:id endpoint", () => {
  const installId = "test-id";
  const orgId = "test-id";
  test("200 status", async () => {
    const { data: spec } = await getInstallById({ installId, orgId });
    expect(spec).toHaveProperty("id");
    expect(spec).toHaveProperty("name");
  }, 60000);

  test.each(badResponseCodes)("%s status", async (code) => {
    const { error, status } = await getInstallById({ installId, orgId });
    expect(status).toBe(code);
    expect(error).toMatchSnapshot({
      error: expect.any(String),
      description: expect.any(String),
      user_error: expect.any(Boolean),
    });
  });
});
