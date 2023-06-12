import { deleteSecrets } from "./delete-secrets";

const mockSecretServiceClient = {
  deleteSecrets: jest
    .fn()
    .mockImplementationOnce((req, cb) => {
      cb(undefined, {
        toObject: jest.fn().mockReturnValue({}),
      });
    })
    .mockImplementationOnce((req, cb) => {
      const err = new Error("some error");
      cb(err);
    }),
};

const mockClients = {
  instanceStatus: mockSecretServiceClient,
};

const mockSecretsIdsInput = {
  appId: "app-123",
  componentId: "component-123",
  installId: "install-123",
  orgId: "org-123",
  secretId: "id-123",
};

test("deleteSecrets resolver should return true on successful mutation", async () => {
  const spec = await deleteSecrets(
    undefined,
    { input: [mockSecretsIdsInput] },
    { clients: mockClients }
  );

  expect(spec).toEqual(true);
});

test("deleteSecrets resolver should return error on failed query", async () => {
  await expect(
    deleteSecrets(undefined, { input: [] }, { clients: mockClients })
  ).rejects.toThrowErrorMatchingInlineSnapshot(`"some error"`);
});

test("deleteSecrets resolver should return error if service client doesn't exist", async () => {
  await expect(
    deleteSecrets(undefined, { input: [] }, { clients: {} })
  ).rejects.toThrowErrorMatchingInlineSnapshot(`"Service isn't available"`);
});
