import { components } from "./components";

const mockDateTimeObject = {
  day: 31,
  hours: 8,
  minutes: 15,
  month: 12,
  seconds: 30,
  year: 1999,
};

const mockComponent = {
  createdAt: mockDateTimeObject,
  id: "test-id",
  name: "test-node",
  updatedAt: mockDateTimeObject,
};

const mockComponentServiceClient = {
  getComponentsByApp: jest
    .fn()
    .mockImplementationOnce((req, cb) => {
      cb(undefined, {
        toObject: jest.fn().mockReturnValue({
          componentsList: [mockComponent],
        }),
      });
    })
    .mockImplementationOnce((req, cb) => {
      cb(undefined, {
        toObject: jest.fn().mockReturnValue({
          componentsList: [],
        }),
      });
    })
    .mockImplementationOnce((req, cb) => {
      const err = new Error("some error");
      cb(err);
    }),
};

const mockClients = {
  component: mockComponentServiceClient,
};

test("components resolver should return Component connection", async () => {
  const spec = await components(
    undefined,
    { appId: "test-id" },
    { clients: mockClients }
  );

  expect(spec).toEqual({
    edges: [
      {
        cursor: "test-id",
        node: {
          config: {
            __typename: "ComponentConfig",
            buildConfig: null,
            deployConfig: null,
          },
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

test("components resolver should return empty connection", async () => {
  const spec = await components(
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

test("components resolver should return error on failed query", async () => {
  await expect(
    components(undefined, { appId: "test-id" }, { clients: mockClients })
  ).rejects.toThrowErrorMatchingInlineSnapshot(`"some error"`);
});

test("components resolver should return error if service client doesn't exist", async () => {
  await expect(
    components(undefined, { appId: "test-id" }, { clients: {} })
  ).rejects.toThrowErrorMatchingInlineSnapshot(`"Service isn't available"`);
});
