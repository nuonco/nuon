import { createDeployment } from "./create-deployment";

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
  createDeployment: jest
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

test("createDeployment resolver should return Deployment object on successful mutation", async () => {
  const spec = await createDeployment(
    undefined,
    { componentId: "component-id" },
    { clients: mockClients }
  );

  expect(spec).toEqual({
    createdAt: "1999-12-31T08:15:30.000+00:00",
    id: "test-id",
    updatedAt: "1999-12-31T08:15:30.000+00:00",
  });
});

test("createDeployment resolver should return error on failed query", async () => {
  await expect(
    createDeployment(undefined, { componentId: "" }, { clients: mockClients })
  ).rejects.toThrowErrorMatchingInlineSnapshot(`"some error"`);
});

test("createDeployment resolver should return error if service client doesn't exist", async () => {
  await expect(
    createDeployment(undefined, { componentId: "" }, { clients: {} })
  ).rejects.toThrowErrorMatchingInlineSnapshot(`"Service isn't available"`);
});
