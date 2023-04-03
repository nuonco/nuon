import { upsertInstall } from "./upsert-install";

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
  upsertInstall: jest
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

test("upsertInstall resolver should return Install object on successful mutation", async () => {
  const spec = await upsertInstall(
    undefined,
    {
      input: {
        appId: "app-id",
        awsSettings: { region: "US_EAST_1", role: "test:role" },
        name: "test-node",
      },
    },
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

test("upsertInstall resolver should return error on failed query", async () => {
  await expect(
    upsertInstall(
      undefined,
      {
        input: {
          appId: "app-id",
          awsSettings: { region: "US_EAST_1", role: "test:role" },
          name: "test-node",
        },
      },
      { clients: mockClients }
    )
  ).rejects.toThrowErrorMatchingInlineSnapshot(`"some error"`);
});

test("upsertInstall resolver should return error if service client doesn't exist", async () => {
  await expect(
    upsertInstall(
      undefined,
      {
        input: {
          appId: "app-id",
          awsSettings: { region: "US_EAST_1", role: "test:role" },
          name: "test-node",
        },
      },
      { clients: {} }
    )
  ).rejects.toThrowErrorMatchingInlineSnapshot(`"Service isn't available"`);
});
