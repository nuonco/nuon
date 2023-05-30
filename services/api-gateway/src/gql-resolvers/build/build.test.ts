import { build } from "./build";

const mockDateTimeObject = {
  day: 31,
  hours: 8,
  minutes: 15,
  month: 12,
  seconds: 30,
  year: 1999,
};

const mockBuild = {
  componentId: "test-component-id",
  createdAt: mockDateTimeObject,
  gitRef: "test-git-ref",
  id: "test-id",
  updatedAt: mockDateTimeObject,
};

const mockBuildServiceClient = {
  getBuild: jest
    .fn()
    .mockImplementationOnce((req, cb) => {
      cb(undefined, {
        toObject: jest.fn().mockReturnValue({
          build: mockBuild,
        }),
      });
    })
    .mockImplementationOnce((req, cb) => {
      const err = new Error("some error");
      cb(err);
    }),
};

const mockClients = {
  build: mockBuildServiceClient,
};

test("build resolver should return a build's status information", async () => {
  const spec = await build(undefined, { id: "123" }, { clients: mockClients });

  expect(spec).toEqual({
    componentId: "test-component-id",
    createdAt: "1999-12-31T08:15:30.000Z",
    gitRef: "test-git-ref",
    id: "test-id",
    updatedAt: "1999-12-31T08:15:30.000Z",
  });
});

test("builds resolver should return error on failed query", async () => {
  await expect(
    build(undefined, { id: "123" }, { clients: mockClients })
  ).rejects.toThrowErrorMatchingInlineSnapshot(`"some error"`);
});

test("builds resolver should return error if service client doesn't exist", async () => {
  await expect(
    build(undefined, { id: "123" }, { clients: {} })
  ).rejects.toThrowErrorMatchingInlineSnapshot(`"Service isn't available"`);
});
