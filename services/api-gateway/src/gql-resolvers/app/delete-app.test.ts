import { deleteApp } from "./delete-app";

const mockAppServiceClient = {
  deleteApp: jest
    .fn()
    .mockImplementationOnce((req, cb) => {
      cb(undefined, {
        toObject: jest.fn().mockReturnValue({
          deleted: true,
        }),
      });
    })
    .mockImplementationOnce((req, cb) => {
      const err = new Error("some error");
      cb(err);
    }),
};

const mockClients = {
  app: mockAppServiceClient,
};

test("deleteApp resolver should return true on successful mutation", async () => {
  const spec = await deleteApp(
    undefined,
    { id: "test-id" },
    { clients: mockClients }
  );

  expect(spec).toBeTruthy();
});

test("deleteApp resolver should return error on failed query", async () => {
  await expect(
    deleteApp(undefined, { id: "test-id" }, { clients: mockClients })
  ).rejects.toThrowErrorMatchingInlineSnapshot(`"some error"`);
});

test("deleteApp resolver should return error if service client doesn't exist", async () => {
  await expect(
    deleteApp(undefined, { id: "test-id" }, { clients: {} })
  ).rejects.toThrowErrorMatchingInlineSnapshot(`"Service isn't available"`);
});
