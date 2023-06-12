import { instances } from "./instances";

const mockDateTimeObject = {
  day: 31,
  hours: 8,
  minutes: 15,
  month: 12,
  seconds: 30,
  year: 1999,
};

const mockInstance = {
  buildId: "test-build-id",
  componentId: "test-component-id",
  createdAt: mockDateTimeObject,
  deployId: "test-deploy-id",
  id: "test-id",
  updatedAt: mockDateTimeObject,
};

const mockInstanceServiceClient = {
  getInstancesByInstall: jest
    .fn()
    .mockImplementationOnce((req, cb) => {
      cb(undefined, {
        toObject: jest.fn().mockReturnValue({
          instancesList: [mockInstance],
        }),
      });
    })
    .mockImplementationOnce((req, cb) => {
      cb(undefined, {
        toObject: jest.fn().mockReturnValue({
          instancesList: [],
        }),
      });
    })
    .mockImplementationOnce((req, cb) => {
      const err = new Error("some error");
      cb(err);
    }),
};

const mockClients = {
  instance: mockInstanceServiceClient,
};

test("instances resolver should return list of instances", async () => {
  const spec = await instances(
    undefined,
    { installId: "123" },
    { clients: mockClients }
  );

  expect(spec).toEqual([
    {
      buildId: "test-build-id",
      componentId: "test-component-id",
      createdAt: "1999-12-31T08:15:30.000Z",
      deployId: "test-deploy-id",
      id: "test-id",
      updatedAt: "1999-12-31T08:15:30.000Z",
    },
  ]);
});

test("instances resolver should return empty array", async () => {
  const spec = await instances(
    undefined,
    { installId: "123" },
    { clients: mockClients }
  );

  expect(spec).toEqual([]);
});

test("instances resolver should return error on failed query", async () => {
  await expect(
    instances(undefined, { installId: "123" }, { clients: mockClients })
  ).rejects.toThrowErrorMatchingInlineSnapshot(`"some error"`);
});

test("instances resolver should return error if service client doesn't exist", async () => {
  await expect(
    instances(undefined, { installId: "123" }, { clients: {} })
  ).rejects.toThrowErrorMatchingInlineSnapshot(`"Service isn't available"`);
});
