import { deployments } from "./deployments";

const mockDateTimeObject = {
  day: 31,
  hours: 8,
  minutes: 15,
  month: 12,
  seconds: 30,
  year: 1999,
};

const mockDeployment = {
  createdAt: mockDateTimeObject,
  id: "test-id",
  updatedAt: mockDateTimeObject,
};

const mockDeploymentServiceClient = {
  getDeploymentsByApps: jest
    .fn()
    .mockImplementationOnce((req, cb) => {
      cb(undefined, {
        toObject: jest.fn().mockReturnValue({
          deploymentsList: [mockDeployment],
        }),
      });
    })
    .mockImplementationOnce((req, cb) => {
      cb(undefined, {
        toObject: jest.fn().mockReturnValue({
          deploymentsList: [],
        }),
      });
    })
    .mockImplementationOnce((req, cb) => {
      const err = new Error("some error");
      cb(err);
    }),
  getDeploymentsByComponents: jest
    .fn()
    .mockImplementationOnce((req, cb) => {
      cb(undefined, {
        toObject: jest.fn().mockReturnValue({
          deploymentsList: [mockDeployment],
        }),
      });
    })
    .mockImplementationOnce((req, cb) => {
      cb(undefined, {
        toObject: jest.fn().mockReturnValue({
          deploymentsList: [],
        }),
      });
    })
    .mockImplementationOnce((req, cb) => {
      const err = new Error("some error");
      cb(err);
    }),
  getDeploymentsByInstalls: jest
    .fn()
    .mockImplementationOnce((req, cb) => {
      cb(undefined, {
        toObject: jest.fn().mockReturnValue({
          deploymentsList: [mockDeployment],
        }),
      });
    })
    .mockImplementationOnce((req, cb) => {
      cb(undefined, {
        toObject: jest.fn().mockReturnValue({
          deploymentsList: [],
        }),
      });
    })
    .mockImplementationOnce((req, cb) => {
      const err = new Error("some error");
      cb(err);
    }),
};

const mockClients = {
  deployment: mockDeploymentServiceClient,
};

test("deployments resolver should return Deployment connection for appIds", async () => {
  const spec = await deployments(
    undefined,
    { appIds: ["test-id"] },
    { clients: mockClients }
  );

  expect(spec).toEqual({
    edges: [
      {
        cursor: "test-id",
        node: {
          createdAt: "1999-12-31T08:15:30.000",
          id: "test-id",
          updatedAt: "1999-12-31T08:15:30.000",
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

test("deployments resolver should return Deployment connection for componentIds", async () => {
  const spec = await deployments(
    undefined,
    { componentIds: ["test-id"] },
    { clients: mockClients }
  );

  expect(spec).toEqual({
    edges: [
      {
        cursor: "test-id",
        node: {
          createdAt: "1999-12-31T08:15:30.000",
          id: "test-id",
          updatedAt: "1999-12-31T08:15:30.000",
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

test("deployments resolver should return Deployment connection for installIds", async () => {
  const spec = await deployments(
    undefined,
    { installIds: ["test-id"] },
    { clients: mockClients }
  );

  expect(spec).toEqual({
    edges: [
      {
        cursor: "test-id",
        node: {
          createdAt: "1999-12-31T08:15:30.000",
          id: "test-id",
          updatedAt: "1999-12-31T08:15:30.000",
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

test("deployments resolver should return empty connection for appIds", async () => {
  const spec = await deployments(
    undefined,
    { appIds: ["test-id"] },
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

test("deployments resolver should return empty connection for componentIds", async () => {
  const spec = await deployments(
    undefined,
    { componentIds: ["test-id"] },
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

test("deployments resolver should return empty connection for installIds", async () => {
  const spec = await deployments(
    undefined,
    { installIds: ["test-id"] },
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

test("deployments resolver should return error on failed query", async () => {
  await expect(
    deployments(
      undefined,
      { componentIds: ["test-id"] },
      { clients: mockClients }
    )
  ).rejects.toThrowErrorMatchingInlineSnapshot(`"some error"`);
});

test("deployments resolver should return error if no query values are provided", async () => {
  await expect(
    deployments(undefined, {}, { clients: mockClients })
  ).rejects.toThrowErrorMatchingInlineSnapshot(
    `"Must provide one of: appIds, componentIds, installIds"`
  );
});

test("deployments resolver should return error if service client doesn't exist", async () => {
  await expect(
    deployments(undefined, { componentIds: ["test-id"] }, { clients: {} })
  ).rejects.toThrowErrorMatchingInlineSnapshot(`"Service isn't available"`);
});
