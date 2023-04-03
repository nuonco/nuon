import { install } from "./install";

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
  getInstall: jest
    .fn()
    .mockImplementationOnce((req, cb) => {
      cb(undefined, {
        toObject: jest.fn().mockReturnValue({
          install: mockInstall,
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

test("install resolver should return install object on successful query", async () => {
  const spec = await install(
    undefined,
    { id: "test-id" },
    { clients: mockClients }
  );

  expect(spec).toEqual({
    createdAt: "1999-12-31T08:15:30.000",
    id: "test-id",
    name: "test-node",
    settings: {
      __typename: "AWSSettings",
      region: "US_EAST_1",
      role: "test:role",
    },
    updatedAt: "1999-12-31T08:15:30.000",
  });
});

test("install resolver should return error on failed query", async () => {
  await expect(
    install(undefined, { id: "test-id" }, { clients: mockClients })
  ).rejects.toThrowErrorMatchingInlineSnapshot(`"some error"`);
});

test("install resolver should return error if service client doesn't exist", async () => {
  await expect(
    install(undefined, { id: "test-id" }, { clients: {} })
  ).rejects.toThrowErrorMatchingInlineSnapshot(`"Service isn't available"`);
});
