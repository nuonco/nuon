import { buildStatus } from "./build-status";

const mockStatus = {
  message: "",
  status: 1,
};

const mockBuildStatusServiceClient = {
  getStatus: jest
    .fn()
    .mockImplementationOnce((req, cb) => {
      cb(undefined, {
        toObject: jest.fn().mockReturnValue(mockStatus),
      });
    })
    .mockImplementationOnce((req, cb) => {
      const err = new Error("some error");
      cb(err);
    }),
};

const mockClients = {
  buildStatus: mockBuildStatusServiceClient,
};

test("buildStatus resolver should return buildStatus object on successful query", async () => {
  const spec = await buildStatus(
    undefined,
    {
      appId: "test-app-id",
      buildId: "test-build-id",
      componentId: "test-component-id",
      orgId: "test-org-id",
    },
    { clients: mockClients }
  );

  expect(spec).toBe("ACTIVE");
});

test("buildStatus resolver should return error on failed query", async () => {
  await expect(
    buildStatus(
      undefined,
      {
        appId: "test-app-id",
        buildId: "test-build-id",
        componentId: "test-component-id",
        orgId: "test-org-id",
      },
      { clients: mockClients }
    )
  ).rejects.toThrowErrorMatchingInlineSnapshot(`"some error"`);
});

test("buildStatus resolver should return error if service client doesn't exist", async () => {
  await expect(
    buildStatus(
      undefined,
      {
        appId: "test-app-id",
        buildId: "test-build-id",
        componentId: "test-component-id",
        orgId: "test-org-id",
      },
      { clients: {} }
    )
  ).rejects.toThrowErrorMatchingInlineSnapshot(`"Service isn't available"`);
});
