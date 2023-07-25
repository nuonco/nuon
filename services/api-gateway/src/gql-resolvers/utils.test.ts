import { WaypointClient } from "../build/waypoint/server/waypoint/main_grpc_pb";
import { getOrgWaypointToken, initOrgWaypointClient } from "./utils";

jest.mock("../build/waypoint/server/waypoint/main_grpc_pb");

const mockClient = WaypointClient as jest.Mocked<typeof WaypointClient>;
const mockOrgWaypointInfo = {
  address: "test.com",
  token: "test",
};
const mockOrgStatusClient = {
  getToken: jest
    .fn()
    .mockImplementationOnce((req, cb) => {
      cb(undefined, {
        toObject: jest.fn().mockReturnValue(mockOrgWaypointInfo),
      });
    })
    .mockImplementationOnce((req, cb) => {
      const err = new Error("some error");
      cb(err);
    }),
};

afterAll(() => {
  jest.restoreAllMocks();
});

test("initOrgWaypointClient should return an initialized waypoint org grpc client", () => {
  initOrgWaypointClient("test.com");
  expect(mockClient).toBeCalledWith("test.com", expect.anything());
});

test("getOrgWaypointToken should resolve a org waypoint token and address", async () => {
  const spec = await getOrgWaypointToken("org-id", mockOrgStatusClient);
  expect(spec).toEqual(mockOrgWaypointInfo);
});

test("getOrgWaypointToken should reject with an error if org waypoint can't be reached", async () => {
  expect(
    getOrgWaypointToken("org-id", mockOrgStatusClient)
  ).rejects.toThrowErrorMatchingInlineSnapshot(`"some error"`);
});
