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

const mockOrg = {
  createdAt: mockDateTimeObject,
  id: "test-id",
  name: "test-node",
  updatedAt: mockDateTimeObject,
};

const request = supertest(
  initServer({
    org: {
      getOrg: jest.fn().mockImplementation((req, cb) => {
        cb(undefined, {
          toObject: jest.fn().mockReturnValue({ org: mockOrg }),
        });
      }),
    },
  })
);

test("Org query should return an org", async () => {
  const spec = await request
    .post("/graphql")
    .set("Authorization", `Bearer ${process.env.TEST_TOKEN}`)
    .send({
      query: `
        query Org($id: ID!) {
          org(id: $id) {
            id
          }
        }
      `,
      variables: {
        id: "org-id",
      },
    })
    .set("Accept", "application/json");

  expect(JSON.parse(spec.text)).toEqual({
    data: {
      org: {
        id: "test-id",
      },
    },
  });
});
