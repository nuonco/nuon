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

const mockInstall = {
  awsSettings: {
    region: 1,
    role: "test:role",
  },
  createdAt: mockDateTimeObject,
  id: "test-id",
  name: "test-node",
  updatedAt: mockDateTimeObject,
};

const request = supertest(initServer({
  install: {
    getInstallsByApp: jest.fn().mockImplementation((req, cb) => {
      cb(undefined, {
        toObject: jest.fn().mockReturnValue({ installsList: [mockInstall] })
      })
    })
  }
}));

test("Installs query should return total count of 1 when a single install is returned", async () => {
  const spec = await request
    .post("/graphql")
    .set("Authorization", `Bearer ${process.env.TEST_TOKEN}`)
    .send({
      query: `
        query Installs($appId: ID!, $options: ConnectionOptions) {
          installs(appId: $appId, options: $options) {
            totalCount
          }
        }
      `,
      variables: {
        appId: "app-id",
        options: {},
      },
    })
    .set("Accept", "application/json");

  expect(JSON.parse(spec.text)).toEqual({
    data: {
      installs: {
        totalCount: 1,
      },
    },
  });
});
