import "dotenv/config";
import supertest from "supertest";
import { initServer } from "../../src/server";

const request = supertest(initServer());

test("createDeployment mutation should return null", async () => {
  const spec = await request
    .post("/graphql")
    .set("Authorization", `Bearer ${process.env.TEST_TOKEN}`)
    .send({
      query: `
        mutation CreateDeployment($componentId: ID!) {
          createDeployment(componentId: $componentId) {
            id
          }
        }
      `,
    })
    .send({ variables: { componentId: "component-id" } })
    .set("Accept", "application/json");

  expect(JSON.parse(spec.text)).toEqual({
    data: {
      createDeployment: null,
    },
  });
});
