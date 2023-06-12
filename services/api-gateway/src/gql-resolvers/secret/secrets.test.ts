import { secrets } from "./secrets";

const mockSecret = {
  id: "test-id",
  key: "test-key",
  value: "test-value",
};

const mockSecretServiceClient = {
  getSecrets: jest
    .fn()
    .mockImplementationOnce((req, cb) => {
      cb(undefined, {
        toObject: jest.fn().mockReturnValue({
          secretsList: [mockSecret],
        }),
      });
    })
    .mockImplementationOnce((req, cb) => {
      cb(undefined, {
        toObject: jest.fn().mockReturnValue({
          secretsList: [],
        }),
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

test("secrets resolver should return list of secrets", async () => {
  const spec = await secrets(
    undefined,
    { appId: "123", componentId: "456", installId: "789", orgId: "000" },
    { clients: mockClients }
  );

  expect(spec).toEqual([
    {
      id: "test-id",
      key: "test-key",
      value: "test-value",
    },
  ]);
});

test("secrets resolver should return empty array", async () => {
  const spec = await secrets(
    undefined,
    { appId: "123", componentId: "456", installId: "789", orgId: "000" },
    { clients: mockClients }
  );

  expect(spec).toEqual([]);
});

test("secrets resolver should return error on failed query", async () => {
  await expect(
    secrets(
      undefined,
      { appId: "123", componentId: "456", installId: "789", orgId: "000" },
      { clients: mockClients }
    )
  ).rejects.toThrowErrorMatchingInlineSnapshot(`"some error"`);
});

test("secrets resolver should return error if service client doesn't exist", async () => {
  await expect(
    secrets(
      undefined,
      { appId: "123", componentId: "456", installId: "789", orgId: "000" },
      { clients: {} }
    )
  ).rejects.toThrowErrorMatchingInlineSnapshot(`"Service isn't available"`);
});
