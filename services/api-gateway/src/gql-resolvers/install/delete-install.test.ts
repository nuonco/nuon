import { deleteInstall } from "./delete-install";

const mockInstallServiceClient = {
  deleteInstall: jest
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
  install: mockInstallServiceClient,
};

test("deleteInstall resolver should return true on successful mutation", async () => {
  const spec = await deleteInstall(
    undefined,
    { id: "test-id" },
    { clients: mockClients }
  );

  expect(spec).toBeTruthy();
});

test("deleteInstall resolver should return error on failed query", async () => {
  await expect(
    deleteInstall(undefined, { id: "test-id" }, { clients: mockClients })
  ).rejects.toThrowErrorMatchingInlineSnapshot(`"some error"`);
});

test("deleteInstall resolver should return error if service client doesn't exist", async () => {
  await expect(
    deleteInstall(undefined, { id: "test-id" }, { clients: {} })
  ).rejects.toThrowErrorMatchingInlineSnapshot(`"Service isn't available"`);
});
