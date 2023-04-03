import { deleteOrg } from "./delete-org";

const mockOrgServiceClient = {
  deleteOrg: jest
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
  org: mockOrgServiceClient,
};

test("deleteOrg resolver should return true on successful mutation", async () => {
  const spec = await deleteOrg(
    undefined,
    { id: "test-id" },
    { clients: mockClients }
  );

  expect(spec).toBeTruthy();
});

test("deleteOrg resolver should return error on failed query", async () => {
  await expect(
    deleteOrg(undefined, { id: "test-id" }, { clients: mockClients })
  ).rejects.toThrowErrorMatchingInlineSnapshot(`"some error"`);
});

test("deleteOrg resolver should return error if service client doesn't exist", async () => {
  await expect(
    deleteOrg(undefined, { id: "test-id" }, { clients: {} })
  ).rejects.toThrowErrorMatchingInlineSnapshot(`"Service isn't available"`);
});
