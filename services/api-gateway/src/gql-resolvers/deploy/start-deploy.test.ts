import { startDeploy } from "./start-deploy";

const mockDeploy = {
  buildId: "test-build-id",
  createdAt: "1999-12-31T08:15:30.000Z",
  id: "test-id",
  installId: "test-install-id",
  updatedAt: "1999-12-31T08:15:30.000Z",
};

const mockDeployServiceClient = {
  startDeploy: jest
    .fn()
    .mockImplementationOnce((req, cb) => {
      cb(undefined, {
        toObject: jest.fn().mockReturnValue({
          deploy: mockDeploy,
        }),
      });
    })
    .mockImplementationOnce((req, cb) => {
      const err = new Error("some error");
      cb(err);
    }),
};

const mockClients = {
  deploy: mockDeployServiceClient,
};

test("startDeploy resolver should return a deploy on successful mutation", async () => {
  const spec = await startDeploy(
    undefined,
    {
      input: {
        buildId: "test-build-id",
        componentId: "test-component-id",
        installId: "test-install-id",
      },
    },
    { clients: mockClients }
  );

  expect(spec).toBeTruthy();
});

test("startDeploy resolver should return error on failed query", async () => {
  await expect(
    startDeploy(
      undefined,
      {
        input: {
          buildId: "test-build-id",
          componentId: "test-component-id",
          installId: "test-install-id",
        },
      },
      { clients: mockClients }
    )
  ).rejects.toThrowErrorMatchingInlineSnapshot(`"some error"`);
});

test("startDeploy resolver should return error if service client doesn't exist", async () => {
  await expect(
    startDeploy(
      undefined,
      {
        input: {
          buildId: "test-build-id",
          componentId: "test-component-id",
          installId: "test-install-id",
        },
      },
      { clients: {} }
    )
  ).rejects.toThrowErrorMatchingInlineSnapshot(`"Service isn't available"`);
});
