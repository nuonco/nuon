import { deploy } from "./deploy";

const mockDateTimeObject = {
  day: 31,
  hours: 8,
  minutes: 15,
  month: 12,
  seconds: 30,
  year: 1999,
};

const mockDeploy = {
  buildId: "test-build-id",
  createdAt: mockDateTimeObject,
  id: "test-id",
  installId: "test-install-id",
  updatedAt: mockDateTimeObject,
};

const mockDeployServiceClient = {
  getDeploy: jest
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

test("deploy resolver should return a single deploy", async () => {
  const spec = await deploy(undefined, { id: "123" }, { clients: mockClients });

  expect(spec).toEqual({
    buildId: "test-build-id",
    createdAt: "1999-12-31T08:15:30.000Z",
    id: "test-id",
    installId: "test-install-id",
    updatedAt: "1999-12-31T08:15:30.000Z",
  });
});

test("deploy resolver should return error on failed query", async () => {
  await expect(
    deploy(undefined, { id: "123" }, { clients: mockClients })
  ).rejects.toThrowErrorMatchingInlineSnapshot(`"some error"`);
});

test("deploy resolver should return error if service client doesn't exist", async () => {
  await expect(
    deploy(undefined, { id: "123" }, { clients: {} })
  ).rejects.toThrowErrorMatchingInlineSnapshot(`"Service isn't available"`);
});
