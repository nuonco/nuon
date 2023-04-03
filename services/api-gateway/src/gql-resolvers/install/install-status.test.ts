import { installStatus } from "./install-status";

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
  installStatus: mockOrgStatusServiceClient,
};

test("installStatus resolver should return installStatus object on successful query", async () => {
  const spec = await installStatus(
    undefined,
    { appId: "test-id", installId: "test-id", orgId: "test-id" },
    { clients: mockClients }
  );

  expect(spec).toBe("ACTIVE");
});

test("installStatus resolver should return error on failed query", async () => {
  await expect(
    installStatus(
      undefined,
      { appId: "test-id", installId: "test-id", orgId: "test-id" },
      { clients: mockClients }
    )
  ).rejects.toThrowErrorMatchingInlineSnapshot(`"some error"`);
});

test("installStatus resolver should return error if service client doesn't exist", async () => {
  await expect(
    installStatus(
      undefined,
      { appId: "test-id", installId: "test-id", orgId: "test-id" },
      { clients: {} }
    )
  ).rejects.toThrowErrorMatchingInlineSnapshot(`"Service isn't available"`);
});
