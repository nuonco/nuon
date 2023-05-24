import { upsertOrg } from "./upsert-org";

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
  upsertOrg: jest
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

test("upsertOrg resolver should return Org object on successful mutation", async () => {
  const spec = await upsertOrg(
    undefined,
    { input: { name: "test-node", ownerId: "owner-id" } },
    { clients: mockClients }
  );

  expect(spec).toEqual({
    createdAt: "1999-12-31T08:15:30.000+00:00",
    id: "test-id",
    name: "test-node",
    updatedAt: "1999-12-31T08:15:30.000+00:00",
  });
});

test("upsertOrg resolver should return error on failed query", async () => {
  await expect(
    upsertOrg(undefined, { input: {} }, { clients: mockClients })
  ).rejects.toThrowErrorMatchingInlineSnapshot(`"some error"`);
});

test("upsertOrg resolver should return error if service client doesn't exist", async () => {
  await expect(
    upsertOrg(undefined, { input: {} }, { clients: {} })
  ).rejects.toThrowErrorMatchingInlineSnapshot(`"Service isn't available"`);
});
