import { deleteComponent } from "./delete-component";

const mockComponentServiceClient = {
  deleteComponent: jest
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
  component: mockComponentServiceClient,
};

test("deleteComponent resolver should return true on successful mutation", async () => {
  const spec = await deleteComponent(
    undefined,
    { id: "test-id" },
    { clients: mockClients }
  );

  expect(spec).toBeTruthy();
});

test("deleteComponent resolver should return error on failed query", async () => {
  await expect(
    deleteComponent(undefined, { id: "test-id" }, { clients: mockClients })
  ).rejects.toThrowErrorMatchingInlineSnapshot(`"some error"`);
});

test("deleteComponent resolver should return error if service client doesn't exist", async () => {
  await expect(
    deleteComponent(undefined, { id: "test-id" }, { clients: {} })
  ).rejects.toThrowErrorMatchingInlineSnapshot(`"Service isn't available"`);
});
