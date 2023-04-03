import { orgStatus } from "./org-status";

const mockStatus = {
  message: "",
  status: 1,
};

const mockOrgStatusServiceClient = {
  getStatus: jest
    .fn()
    .mockImplementationOnce((req, cb) => {
      cb(undefined, {
        toObject: jest.fn().mockReturnValue(mockStatus),
      });
    })
    .mockImplementationOnce((req, cb) => {
      const err = new Error("some error");
      cb(err);
    }),
};

const mockClients = {
  orgStatus: mockOrgStatusServiceClient,
};

test("orgStatus resolver should return orgStatus object on successful query", async () => {
  const spec = await orgStatus(
    undefined,
    { id: "test-id" },
    { clients: mockClients }
  );

  expect(spec).toBe("ACTIVE");
});

test("orgStatus resolver should return error on failed query", async () => {
  await expect(
    orgStatus(undefined, { id: "test-id" }, { clients: mockClients })
  ).rejects.toThrowErrorMatchingInlineSnapshot(`"some error"`);
});

test("orgStatus resolver should return error if service client doesn't exist", async () => {
  await expect(
    orgStatus(undefined, { id: "test-id" }, { clients: {} })
  ).rejects.toThrowErrorMatchingInlineSnapshot(`"Service isn't available"`);
});
