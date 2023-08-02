import express, { Express, Request, Response } from "express";
import {
  errorHandler,
  gqlContextHandler,
  healthzHandler,
  initServer,
  serverListenHandler,
  startServer,
} from "./server";
import { env, logger } from "./utils";

jest.mock("express");
jest.mock("./utils/logger");

const mockExpressFunction = express as jest.MockedFunction<typeof express>;
const mockedExpress: Partial<Express> = {
  get: jest.fn(),
  listen: jest.fn(),
  use: jest.fn(),
};
const mockLogger = logger as jest.Mocked<typeof logger>;
const mockRequest: Partial<Request> = {};
const mockResponse: Partial<Response> = {
  json: jest.fn().mockReturnThis(),
  send: jest.fn().mockReturnThis(),
  status: jest.fn().mockReturnThis(),
};

afterAll(() => {
  jest.restoreAllMocks();
});

beforeEach(() => {
  mockExpressFunction.mockImplementation(() => mockedExpress as Express);
});

test("errorHandler should return an error status code & json with error message", () => {
  const mockError = {
    message: "Something bad happened",
    name: "error",
    statusCode: 401,
  };
  errorHandler(
    mockError,
    mockRequest as Request,
    mockResponse as Response,
    jest.fn()
  );
  expect(mockResponse.status).toBeCalledWith(401);
  expect(mockResponse.json).toBeCalled();
});

test('healthzHandler should return a 200 status code & json with version number & status of "ok"', () => {
  healthzHandler(mockRequest as Request, mockResponse as Response, jest.fn());

  expect(mockResponse.status).toBeCalledWith(200);
  expect(mockResponse.json).toBeCalled();
});

test("gqlContextHandler should return a context object", () => {
  const context = gqlContextHandler({})({ req: mockRequest as Request });

  expect(context).toEqual({ clients: {}, req: mockRequest, user: {} });
});

test("gqlContextHandler should return a context object with a user when request has auth payload", () => {
  const mockAuthedRequest = {
    auth: { payload: { sub: "user-id" } },
    ...mockRequest,
  };
  const context = gqlContextHandler({})({
    req: mockAuthedRequest as Request,
  });

  expect(context).toEqual({
    clients: {},
    req: mockAuthedRequest,
    user: { id: "user-id" },
  });
});

test("initServer should return an Express server with middleware & healthz routes applied", () => {
  const testServer = initServer();

  expect(mockLogger.debug).toBeCalledWith("Initializing server");
  expect(mockExpressFunction).toBeCalled();
  expect(testServer.use).toBeCalledTimes(3);
  expect(testServer.get).toBeCalledTimes(3);
});

test("serverListenHandler should call logger.info with port information", () => {
  serverListenHandler();

  expect(mockLogger.info).toBeCalledWith(`Server running at port ${env.PORT}`);
});

test("startServer should initServer & listen at the configured port", () => {
  startServer();

  expect(mockLogger.debug).toBeCalledWith("Attempting to start server");
  expect(mockedExpress.listen).toBeCalledWith(env.PORT, serverListenHandler);
});
