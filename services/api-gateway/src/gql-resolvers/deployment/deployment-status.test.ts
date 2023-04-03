import { deploymentStatus } from "./deployment-status";

const mockStatus = {
  message: "",
  status: 1,
};

const mockDeploymentStatusServiceClient = {
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
  deploymentStatus: mockDeploymentStatusServiceClient,
};

test("deploymentStatus resolver should return deploymentStatus object on successful query", async () => {
  const spec = await deploymentStatus(
    undefined,
    {
      appId: "test-app-id",
      componentId: "test-component-id",
      deploymentId: "test-deploy-id",
      orgId: "test-org-id",
    },
    { clients: mockClients }
  );

  expect(spec).toBe("ACTIVE");
});

test("deploymentStatus resolver should return error on failed query", async () => {
  await expect(
    deploymentStatus(
      undefined,
      {
        appId: "test-app-id",
        componentId: "test-component-id",
        deploymentId: "test-deploy-id",
        orgId: "test-org-id",
      },
      { clients: mockClients }
    )
  ).rejects.toThrowErrorMatchingInlineSnapshot(`"some error"`);
});

test("deploymentStatus resolver should return error if service client doesn't exist", async () => {
  await expect(
    deploymentStatus(
      undefined,
      {
        appId: "test-app-id",
        componentId: "test-component-id",
        deploymentId: "test-deploy-id",
        orgId: "test-org-id",
      },
      { clients: {} }
    )
  ).rejects.toThrowErrorMatchingInlineSnapshot(`"Service isn't available"`);
});
