import { instanceStatus } from "./instance-status";

const mockInfo = {
  hostname: "test.com",
};

const mockStatus = {
  message: "",
  status: 1,
};

const mockInstanceStatusServiceClient = {
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
  instanceStatus: mockInstanceStatusServiceClient,
};

test("instance resolver should return instance object on successful query", async () => {
  const spec = await instanceStatus(
    undefined,
    {
      appId: "test-id",
      componentId: "test-id",
      deployId: "test-id",
      installId: "test-id",
      orgId: "test-id",
    },
    { clients: mockClients }
  );

  expect(spec).toEqual({
    __typename: "InstanceStatus",
    hostname: "test.com",
    status: "ACTIVE",
  });
});

test("instance resolver should return error if service client doesn't exist", async () => {
  await expect(
    instanceStatus(
      undefined,
      {
        appId: "test-id",
        componentId: "test-id",
        deployId: "test-id",
        installId: "test-id",
        orgId: "test-id",
      },
      { clients: {} }
    )
  ).rejects.toThrowErrorMatchingInlineSnapshot(`"Service isn't available"`);
});
