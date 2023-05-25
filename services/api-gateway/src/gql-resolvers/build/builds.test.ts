import { builds } from "./builds";

const mockBuild = {
  componentId: "test-component-id",
  createdAt: "1999-12-31T08:15:30.000Z",
  gitRef: "test-git-ref",
  id: "test-id",
  updatedAt: "1999-12-31T08:15:30.000Z",
};

const mockBuildServiceClient = {
  queryBuilds: jest
    .fn()
    .mockImplementationOnce((req, cb) => {
      cb(undefined, {
        toObject: jest.fn().mockReturnValue({
          buildsList: [mockBuild],
        }),
      });
    })
    .mockImplementationOnce((req, cb) => {
      cb(undefined, {
        toObject: jest.fn().mockReturnValue({
          buildsList: [],
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

test("builds resolver should return list of builds", async () => {
  const spec = await builds(
    undefined,
    { componentId: "123" },
    { clients: mockClients }
  );

  expect(spec).toEqual([
    {
      componentId: "test-component-id",
      createdAt: "1999-12-31T08:15:30.000Z",
      gitRef: "test-git-ref",
      id: "test-id",
      updatedAt: "1999-12-31T08:15:30.000Z",
    },
  ]);
});

test("builds resolver should return empty array", async () => {
  const spec = await builds(
    undefined,
    { componentId: "123" },
    { clients: mockClients }
  );

  expect(spec).toEqual([]);
});

test("builds resolver should return error on failed query", async () => {
  await expect(
    builds(undefined, { componentId: "123" }, { clients: mockClients })
  ).rejects.toThrowErrorMatchingInlineSnapshot(`"some error"`);
});

test("builds resolver should return error if service client doesn't exist", async () => {
  await expect(
    builds(undefined, { componentId: "123" }, { clients: {} })
  ).rejects.toThrowErrorMatchingInlineSnapshot(`"Service isn't available"`);
});
