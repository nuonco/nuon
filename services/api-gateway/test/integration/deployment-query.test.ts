import "dotenv/config";
import supertest from "supertest";
import { initServer } from "../../src/server";

const request = supertest(initServer());

test("Deployment query should return null", async () => {
  const spec = await request
    .post("/graphql")
    .set("Authorization", `Bearer ${process.env.TEST_TOKEN}`)
    .send({
      query: `
        query Deployment($id: ID!) {
          deployment(id: $id) {
            id
          }
        }
      `,
      variables: {
        id: "deployment-id",
      },
    })
    .set("Accept", "application/json");

  expect(JSON.parse(spec.text)).toEqual({
    data: {
      deployment: null,
    },
  });
});
