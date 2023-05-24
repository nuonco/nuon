import { upsertApp } from "./upsert-app";

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
  upsertApp: jest
    .fn()
    .mockImplementationOnce((req, cb) => {
      cb(undefined, {
        toObject: jest.fn().mockReturnValue({
          app: mockApp,
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

test("upsertApp resolver should return App object on successful mutation", async () => {
  const spec = await upsertApp(
    undefined,
    { input: { name: "test-node", orgId: "org-id" } },
    { clients: mockClients }
  );

  expect(spec).toEqual({
    createdAt: "1999-12-31T08:15:30.000+00:00",
    id: "test-id",
    name: "test-node",
    updatedAt: "1999-12-31T08:15:30.000+00:00",
  });
});

test("upsertApp resolver should return error on failed query", async () => {
  await expect(
    upsertApp(undefined, { input: {} }, { clients: mockClients })
  ).rejects.toThrowErrorMatchingInlineSnapshot(`"some error"`);
});

test("upsertApp resolver should return error if service client doesn't exist", async () => {
  await expect(
    upsertApp(undefined, { input: {} }, { clients: {} })
  ).rejects.toThrowErrorMatchingInlineSnapshot(`"Service isn't available"`);
});
