import { cancelBuild } from "./cancel-build";

const mockBuildServiceClient = {
  cancelBuild: jest
    .fn()
    .mockImplementationOnce((req, cb) => {
      cb(undefined, {
        toObject: jest.fn().mockReturnValue({ deleted: true }),
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

test("cancelBuild resolver should return true on successful mutation", async () => {
  const spec = await cancelBuild(
    undefined,
    { id: "test-id" },
    { clients: mockClients }
  );

  expect(spec).toBeTruthy();
});

test("cancelBuild resolver should return error on failed query", async () => {
  await expect(
    cancelBuild(undefined, { id: "test-id" }, { clients: mockClients })
  ).rejects.toThrowErrorMatchingInlineSnapshot(`"some error"`);
});

test("cancelBuild resolver should return error if service client doesn't exist", async () => {
  await expect(
    cancelBuild(undefined, { id: "test-id" }, { clients: {} })
  ).rejects.toThrowErrorMatchingInlineSnapshot(`"Service isn't available"`);
});
