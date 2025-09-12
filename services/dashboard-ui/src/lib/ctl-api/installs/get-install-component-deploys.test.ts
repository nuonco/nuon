import "@test/mock-auth";
import { badResponseCodes } from "@test/utils";
import { describe, expect, test } from "vitest";
import { getInstallComponentDeploys } from "./get-install-component-deploys";

describe("getInstallComponentDeploys should handle response status codes from GET installs/:installId/components/:componentId/deploys endpoint", () => {
  const installId = "test-install-id";
  const componentId = "test-component-id";
  const orgId = "test-org-id";

  test("200 status with pagination", async () => {
    const { data: deploys } = await getInstallComponentDeploys({ 
      installId, 
      componentId, 
      orgId,
      limit: 10,
      offset: 0
    });
    expect(deploys).toBeInstanceOf(Array);
  }, 60000);

  test.each(badResponseCodes)("%s status", async (code) => {
    const { error, status } = await getInstallComponentDeploys({ 
      installId, 
      componentId, 
      orgId 
    });
    expect(status).toBe(code);
    expect(error).toMatchSnapshot({
      error: expect.any(String),
      description: expect.any(String),
      user_error: expect.any(Boolean),
    });
  });
});
