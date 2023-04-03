import { instance } from "./instance";

const mockInfo = {
  response: {
    hostname: "test.com",
  },
};

const mockStatus = {
  message: "",
  status: 1,
};

const mockOrgStatusServiceClient = {
  getInfo: jest.fn().mockImplementationOnce((req, cb) => {
    cb(undefined, {
      toObject: jest.fn().mockReturnValue(mockInfo),
    });
  }),
  getStatus: jest.fn().mockImplementationOnce((req, cb) => {
    cb(undefined, {
      toObject: jest.fn().mockReturnValue(mockStatus),
    });
  }),
};

const mockClients = {
  instance: mockOrgStatusServiceClient,
};

test("instance resolver should return instance object on successful query", async () => {
  const spec = await instance(
    undefined,
    {
      appId: "test-id",
      componentId: "test-id",
      deploymentId: "test-id",
      installId: "test-id",
      orgId: "test-id",
    },
    { clients: mockClients }
  );

  expect(spec).toEqual({
    __typename: "Instance",
    hostname: "test.com",
    status: "ACTIVE",
  });
});

test("instance resolver should return error if service client doesn't exist", async () => {
  await expect(
    instance(
      undefined,
      {
        appId: "test-id",
        componentId: "test-id",
        deploymentId: "test-id",
        installId: "test-id",
        orgId: "test-id",
      },
      { clients: {} }
    )
  ).rejects.toThrowErrorMatchingInlineSnapshot(`"Service isn't available"`);
});
