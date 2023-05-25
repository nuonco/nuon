import { installs } from "./installs";

const mockDateTimeObject = {
  day: 31,
  hours: 8,
  minutes: 15,
  month: 12,
  seconds: 30,
  year: 1999,
};

const mockInstall = {
  awsSettings: {
    region: 1,
    role: "test:role",
  },
  createdAt: mockDateTimeObject,
  id: "test-id",
  name: "test-node",
  updatedAt: mockDateTimeObject,
};

const mockInstallServiceClient = {
  getInstallsByApp: jest
    .fn()
    .mockImplementationOnce((req, cb) => {
      cb(undefined, {
        toObject: jest.fn().mockReturnValue({
          installsList: [mockInstall],
        }),
      });
    })
    .mockImplementationOnce((req, cb) => {
      cb(undefined, {
        toObject: jest.fn().mockReturnValue({
          installsList: [],
        }),
      });
    })
    .mockImplementationOnce((req, cb) => {
      const err = new Error("some error");
      cb(err);
    }),
};

const mockClients = {
  install: mockInstallServiceClient,
};

test("installs resolver should return Install connection", async () => {
  const spec = await installs(
    undefined,
    { appId: "test-id" },
    { clients: mockClients }
  );

  expect(spec).toEqual({
    edges: [
      {
        cursor: "test-id",
        node: {
          createdAt: "1999-12-31T08:15:30.000Z",
          id: "test-id",
          name: "test-node",
          settings: {
            __typename: "AWSSettings",
            region: "US_EAST_1",
            role: "test:role",
          },
          updatedAt: "1999-12-31T08:15:30.000Z",
        },
      },
    ],
    pageInfo: {
      endCursor: null,
      hasNextPage: false,
      hasPreviousPage: false,
      startCursor: null,
    },
    totalCount: 1,
  });
});

test("installs resolver should return empty connection", async () => {
  const spec = await installs(
    undefined,
    { appId: "test-id" },
    { clients: mockClients }
  );

  expect(spec).toEqual({
    edges: [],
    pageInfo: {
      endCursor: null,
      hasNextPage: false,
      hasPreviousPage: false,
      startCursor: null,
    },
    totalCount: 0,
  });
});

test("installs resolver should return error on failed query", async () => {
  await expect(
    installs(undefined, { appId: "test-id" }, { clients: mockClients })
  ).rejects.toThrowErrorMatchingInlineSnapshot(`"some error"`);
});

test("installs resolver should return error if service client doesn't exist", async () => {
  await expect(
    installs(undefined, { appId: "test-id" }, { clients: {} })
  ).rejects.toThrowErrorMatchingInlineSnapshot(`"Service isn't available"`);
});
