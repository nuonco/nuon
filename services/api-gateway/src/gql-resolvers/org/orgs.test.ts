import { orgs } from "./orgs";

const mockDateTimeObject = {
  day: 31,
  hours: 8,
  minutes: 15,
  month: 12,
  seconds: 30,
  year: 1999,
};

const mockOrg = {
  createdAt: mockDateTimeObject,
  id: "test-id",
  name: "test-node",
  updatedAt: mockDateTimeObject,
};

const mockOrgServiceClient = {
  getOrgsByMember: jest
    .fn()
    .mockImplementationOnce((req, cb) => {
      cb(undefined, {
        toObject: jest.fn().mockReturnValue({
          orgsList: [mockOrg],
        }),
      });
    })
    .mockImplementationOnce((req, cb) => {
      cb(undefined, {
        toObject: jest.fn().mockReturnValue({
          orgsList: [],
        }),
      });
    })
    .mockImplementationOnce((req, cb) => {
      const err = new Error("some error");
      cb(err);
    }),
};

const mockClients = {
  org: mockOrgServiceClient,
};

test("orgs resolver should return Org connection", async () => {
  const spec = await orgs(
    undefined,
    { memberId: "test-id" },
    { clients: mockClients }
  );

  expect(spec).toEqual({
    edges: [
      {
        cursor: "test-id",
        node: {
          createdAt: "1999-12-31T08:15:30.000",
          id: "test-id",
          name: "test-node",
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

test("orgs resolver should return empty connection", async () => {
  const spec = await orgs(
    undefined,
    { memberId: "test-id" },
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

test("orgs resolver should return error on failed query", async () => {
  await expect(
    orgs(undefined, { memberId: "test-id" }, { clients: mockClients })
  ).rejects.toThrowErrorMatchingInlineSnapshot(`"some error"`);
});

test("orgs resolver should return error if service client doesn't exist", async () => {
  await expect(
    orgs(undefined, { memberId: "test-id" }, { clients: {} })
  ).rejects.toThrowErrorMatchingInlineSnapshot(`"Service isn't available"`);
});
