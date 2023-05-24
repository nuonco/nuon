import { deployment } from "./deployment";

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
  getDeployment: jest
    .fn()
    .mockImplementationOnce((req, cb) => {
      cb(undefined, {
        toObject: jest.fn().mockReturnValue({
          deployment: mockDeployment,
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

test("deployment resolver should return deployment object on successful query", async () => {
  const spec = await deployment(
    undefined,
    { id: "test-id" },
    { clients: mockClients }
  );

  expect(spec).toEqual({
    createdAt: "1999-12-31T08:15:30.000+00:00",
    id: "test-id",
    updatedAt: "1999-12-31T08:15:30.000+00:00",
  });
});

test("deployment resolver should return error on failed query", async () => {
  await expect(
    deployment(undefined, { id: "test-id" }, { clients: mockClients })
  ).rejects.toThrowErrorMatchingInlineSnapshot(`"some error"`);
});

test("deployment resolver should return error if service client doesn't exist", async () => {
  await expect(
    deployment(undefined, { id: "test-id" }, { clients: {} })
  ).rejects.toThrowErrorMatchingInlineSnapshot(`"Service isn't available"`);
});
