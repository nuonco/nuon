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
      getOrgsByMember: jest.fn().mockImplementation((req, cb) => {
        cb(undefined, {
          toObject: jest.fn().mockReturnValue({ orgsList: [mockOrg] }),
        });
      }),
    },
  })
);

test("Orgs query should return total count of 0", async () => {
  const spec = await request
    .post("/graphql")
    .set("Authorization", `Bearer ${process.env.TEST_TOKEN}`)
    .send({
      query: `
        query Orgs($memberId: ID!, $options: ConnectionOptions) {
          orgs(memberId: $memberId, options: $options) {
            totalCount
          }
        }
      `,
      variables: {
        memberId: "user-id",
      },
    })
    .set("Accept", "application/json");

  expect(JSON.parse(spec.text)).toEqual({
    data: {
      orgs: {
        totalCount: 1,
      },
    },
  });
});
