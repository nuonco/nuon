import { component } from "./component";

const mockDateTimeObject = {
  day: 31,
  hours: 8,
  minutes: 15,
  month: 12,
  seconds: 30,
  year: 1999,
};

const mockComponent = {
  createdAt: mockDateTimeObject,
  id: "test-id",
  name: "test-node",
  updatedAt: mockDateTimeObject,
};

const mockComponentServiceClient = {
  getComponent: jest
    .fn()
    .mockImplementationOnce((req, cb) => {
      cb(undefined, {
        toObject: jest.fn().mockReturnValue({
          component: mockComponent,
        }),
      });
    })
    .mockImplementationOnce((req, cb) => {
      const err = new Error("some error");
      cb(err);
    }),
};

const mockClients = {
  component: mockComponentServiceClient,
};

test("component resolver should return component object on successful query", async () => {
  const spec = await component(
    undefined,
    { id: "test-id" },
    { clients: mockClients }
  );

  expect(spec).toEqual({
    config: {
      __typename: "ComponentConfig",
      buildConfig: null,
      deployConfig: null,
    },
    createdAt: "1999-12-31T08:15:30.000Z",
    id: "test-id",
    name: "test-node",
    updatedAt: "1999-12-31T08:15:30.000Z",
  });
});

test("component resolver should return error on failed query", async () => {
  await expect(
    component(undefined, { id: "test-id" }, { clients: mockClients })
  ).rejects.toThrowErrorMatchingInlineSnapshot(`"some error"`);
});

test("component resolver should return error if service client doesn't exist", async () => {
  await expect(
    component(undefined, { id: "test-id" }, { clients: {} })
  ).rejects.toThrowErrorMatchingInlineSnapshot(`"Service isn't available"`);
});
