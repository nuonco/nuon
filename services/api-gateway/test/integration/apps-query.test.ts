import "dotenv/config";
import supertest from "supertest";
import { initServer } from "../../src/server";

const mockDateTimeObject = {
  day: 31,
  hours: 8,
  minutes: 15,
  month: 12,
  seconds: 30,
  year: 1999,
};

const mockApp = {
  createdAt: mockDateTimeObject,
  id: "test-id",
  name: "test-node",
  updatedAt: mockDateTimeObject,
};

const request = supertest(
  initServer({
    app: {
      getAppsByOrg: jest.fn().mockImplementation((req, cb) => {
        cb(undefined, {
          toObject: jest.fn().mockReturnValue({ appsList: [mockApp] }),
        });
      }),
    },
  })
);

test("Apps query should return total count of 0", async () => {
  const spec = await request
    .post("/graphql")
    .set("Authorization", `Bearer ${process.env.TEST_TOKEN}`)
    .send({
      query: `
        query Apps($orgId: ID!, $options: ConnectionOptions) {
          apps(orgId: $orgId, options: $options) {
            totalCount
          }
        }
      `,
      variables: {
        orgId: "user-id",
      },
    })
    .set("Accept", "application/json");

  expect(JSON.parse(spec.text)).toEqual({
    data: {
      apps: {
        totalCount: 1,
      },
    },
  });
});
