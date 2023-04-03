import { app } from "./app";

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
  getApp: jest
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

test("app resolver should return app object on successful query", async () => {
  const spec = await app(
    undefined,
    { id: "test-id" },
    { clients: mockClients }
  );

  expect(spec).toEqual({
    createdAt: "1999-12-31T08:15:30.000",
    id: "test-id",
    name: "test-node",
    updatedAt: "1999-12-31T08:15:30.000",
  });
});

test("app resolver should return error on failed query", async () => {
  await expect(
    app(undefined, { id: "test-id" }, { clients: mockClients })
  ).rejects.toThrowErrorMatchingInlineSnapshot(`"some error"`);
});

test("app resolver should return error if service client doesn't exist", async () => {
  await expect(
    app(undefined, { id: "test-id" }, { clients: {} })
  ).rejects.toThrowErrorMatchingInlineSnapshot(`"Service isn't available"`);
});
