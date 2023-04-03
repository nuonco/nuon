import { initServiceClients } from "./grpc-clients";

test("initServiceClients should return a map of initialized gRPC clients", () => {
  const mockClientClass = jest.fn().mockImplementation(() => ({}));
  const mockServices = [{ name: "test", url: "localhost:8080" }];
  const mockClients = { test: mockClientClass };

  const emptyClients = initServiceClients([], {});
  expect(emptyClients).toEqual({});

  const missingClients = initServiceClients(
    [{ name: "noop", url: "localhost:8080" }],
    { test: mockClientClass }
  );
  expect(missingClients).toEqual({});
  expect(mockClientClass).not.toBeCalled();

  const clients = initServiceClients(mockServices, mockClients);
  expect(clients).toEqual({
    test: {},
  });
  expect(mockClientClass).toBeCalled();
});
