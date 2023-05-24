import { apps } from "./apps";

const mockDateTimeObject = {
  day: 31,
  hours: 8,
  minutes: 15,
  month: 12,
  seconds: 30,
  year: 1999,
};

const mockApp = {
  createdAt: mockDateTimeObject,
  id: "test-id",
  name: "test-node",
  updatedAt: mockDateTimeObject,
};

const mockAppServiceClient = {
  getAppsByOrg: jest
    .fn()
    .mockImplementationOnce((req, cb) => {
      cb(undefined, {
        toObject: jest.fn().mockReturnValue({
          appsList: [mockApp],
        }),
      });
    })
    .mockImplementationOnce((req, cb) => {
      cb(undefined, {
        toObject: jest.fn().mockReturnValue({
          appsList: [],
        }),
      });
    })
    .mockImplementationOnce((req, cb) => {
      const err = new Error("some error");
      cb(err);
    }),
};

const mockClients = {
  app: mockAppServiceClient,
};

test("apps resolver should return App connection", async () => {
  const spec = await apps(
    undefined,
    { orgId: "test-id" },
    { clients: mockClients }
  );

  expect(spec).toEqual({
    edges: [
      {
        cursor: "test-id",
        node: {
          createdAt: "1999-12-31T08:15:30.000+00:00",
          id: "test-id",
          name: "test-node",
          updatedAt: "1999-12-31T08:15:30.000+00:00",
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

test("apps resolver should return empty connection", async () => {
  const spec = await apps(
    undefined,
    { orgId: "test-id" },
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

test("apps resolver should return error on failed query", async () => {
  await expect(
    apps(undefined, { orgId: "test-id" }, { clients: mockClients })
  ).rejects.toThrowErrorMatchingInlineSnapshot(`"some error"`);
});

test("apps resolver should return error if service client doesn't exist", async () => {
  await expect(
    apps(undefined, { orgId: "test-id" }, { clients: {} })
  ).rejects.toThrowErrorMatchingInlineSnapshot(`"Service isn't available"`);
});
