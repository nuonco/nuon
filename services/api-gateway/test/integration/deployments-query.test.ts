import "dotenv/config";
import supertest from "supertest";
import { initServer } from "../../src/server";

const request = supertest(initServer());

test("Deployments query should return total count of 0", async () => {
  const spec = await request
    .post("/graphql")
    .set("Authorization", `Bearer ${process.env.TEST_TOKEN}`)
    .send({
      query: `
        query Deployments($componentIds: [ID!]!, $options: ConnectionOptions) {
          deployments(componentIds: $componentIds, options: $options) {
            totalCount
          }
        }
      `,
      variables: {
        componentIds: ["component-id"],
        options: {},
      },
    })
    .set("Accept", "application/json");

  expect(JSON.parse(spec.text)).toEqual({
    data: {
      deployments: {
        totalCount: 0,
      },
    },
  });
});
