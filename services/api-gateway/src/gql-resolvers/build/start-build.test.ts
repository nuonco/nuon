import { startBuild } from "./start-build";

const mockBuild = {
  componentId: "test-component-id",
  createdAt: "1999-12-31T08:15:30.000",
  gitRef: "test-git-ref",
  id: "test-id",
  updatedAt: "1999-12-31T08:15:30.000",
};

const mockBuildServiceClient = {
  startBuild: jest
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

test("startBuild resolver should return true on successful mutation", async () => {
  const spec = await startBuild(
    undefined,
    { input: { componentId: "test-component-id", gitRef: "test-git-ref" } },
    { clients: mockClients }
  );

  expect(spec).toBeTruthy();
});

test("startBuild resolver should return error on failed query", async () => {
  await expect(
    startBuild(
      undefined,
      { input: { componentId: "test-component-id", gitRef: "test-git-ref" } },
      { clients: mockClients }
    )
  ).rejects.toThrowErrorMatchingInlineSnapshot(`"some error"`);
});

test("startBuild resolver should return error if service client doesn't exist", async () => {
  await expect(
    startBuild(
      undefined,
      { input: { componentId: "test-component-id", gitRef: "test-git-ref" } },
      { clients: {} }
    )
  ).rejects.toThrowErrorMatchingInlineSnapshot(`"Service isn't available"`);
});
