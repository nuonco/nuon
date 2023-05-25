import { org } from "./org";

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
  getOrg: jest
    .fn()
    .mockImplementationOnce((req, cb) => {
      cb(undefined, {
        toObject: jest.fn().mockReturnValue({
          org: mockOrg,
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

test("org resolver should return org object on successful query", async () => {
  const spec = await org(
    undefined,
    { id: "test-id" },
    { clients: mockClients }
  );

  expect(spec).toEqual({
    createdAt: "1999-12-31T08:15:30.000Z",
    id: "test-id",
    name: "test-node",
    updatedAt: "1999-12-31T08:15:30.000Z",
  });
});

test("org resolver should return error on failed query", async () => {
  await expect(
    org(undefined, { id: "test-id" }, { clients: mockClients })
  ).rejects.toThrowErrorMatchingInlineSnapshot(`"some error"`);
});

test("org resolver should return error if service client doesn't exist", async () => {
  await expect(
    org(undefined, { id: "test-id" }, { clients: {} })
  ).rejects.toThrowErrorMatchingInlineSnapshot(`"Service isn't available"`);
});
