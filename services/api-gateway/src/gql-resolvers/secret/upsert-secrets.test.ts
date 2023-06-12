import { parseSecretInput, upsertSecrets } from "./upsert-secrets";

const mockSecretServiceClient = {
  upsertSecrets: jest
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

const mockNewSecret = {
  appId: "app-123",
  componentId: "component-123",
  installId: "install-123",
  orgId: "org-123",
  secrets: [
    {
      key: "key-1",
      value: "value-1",
    },
  ],
};

const mockUpdatedSecrets = {
  appId: "app-123",
  componentId: "component-123",
  installId: "install-123",
  orgId: "org-123",
  secrets: [
    {
      id: "id-1",
      key: "key-1",
      value: "value-1",
    },
    {
      id: "id-2",
      key: "key-2",
      value: "value-2",
    },
  ],
};

test("parseSecretInput should return array with one secret", () => {
  const spec = parseSecretInput([mockNewSecret]);
  expect(spec[0].toObject()).toMatchInlineSnapshot(`
    {
      "appId": "app-123",
      "componentId": "component-123",
      "id": "",
      "installId": "install-123",
      "key": "key-1",
      "orgId": "org-123",
      "value": "value-1",
    }
  `);
});

test("parseSecretInput should return array with more than one secrets", () => {
  const spec = parseSecretInput([mockUpdatedSecrets]);
  expect(spec[0].toObject()).toMatchInlineSnapshot(`
    {
      "appId": "app-123",
      "componentId": "component-123",
      "id": "id-1",
      "installId": "install-123",
      "key": "key-1",
      "orgId": "org-123",
      "value": "value-1",
    }
  `);
  expect(spec[1].toObject()).toMatchInlineSnapshot(`
    {
      "appId": "app-123",
      "componentId": "component-123",
      "id": "id-2",
      "installId": "install-123",
      "key": "key-2",
      "orgId": "org-123",
      "value": "value-2",
    }
  `);
});

test("upsertSecrets mutation should return no result for successful execution", async () => {
  const spec = await upsertSecrets(
    undefined,
    {
      input: [mockUpdatedSecrets],
    },
    { clients: mockClients }
  );

  expect(spec).toEqual(true);
});

test("upsertSecrets mutation should return error on failed query", async () => {
  await expect(
    upsertSecrets(undefined, { input: [] }, { clients: mockClients })
  ).rejects.toThrowErrorMatchingInlineSnapshot(`"some error"`);
});

test("upsertSecrets mutation should return error if service client doesn't exist", async () => {
  await expect(
    upsertSecrets(undefined, { input: [] }, { clients: {} })
  ).rejects.toThrowErrorMatchingInlineSnapshot(`"Service isn't available"`);
});
