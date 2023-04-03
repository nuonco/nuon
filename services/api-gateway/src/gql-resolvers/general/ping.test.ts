import { ping } from "./ping";

const mockStatusServiceClient = {
  ping: jest
    .fn()
    .mockImplementationOnce((req, cb) => {
      cb(undefined, { getStatus: jest.fn().mockReturnValue("ok") });
    })
    .mockImplementationOnce((req, cb) => {
      const err = new Error("some error");
      cb(err);
    }),
};

const mockClients = {
  status: mockStatusServiceClient,
};

test("ping resolver should return ok on successful query", async () => {
  const spec = await ping(undefined, undefined, { clients: mockClients });

  expect(spec).toBe("ok");
});

test("ping resolver should return error on failed query", async () => {
  await expect(
    ping(undefined, undefined, { clients: mockClients })
  ).rejects.toThrowErrorMatchingInlineSnapshot(`"some error"`);
});

test("ping resolver should return error if service client doesn't exist", async () => {
  await expect(
    ping(undefined, undefined, { clients: {} })
  ).rejects.toThrowErrorMatchingInlineSnapshot(`"Service isn't available"`);
});
