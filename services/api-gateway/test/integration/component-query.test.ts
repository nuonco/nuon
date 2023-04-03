import "dotenv/config";
import supertest from "supertest";
import { initServer } from "../../src/server";

const request = supertest(initServer());

test("Component query should return null", async () => {
  const spec = await request
    .post("/graphql")
    .set("Authorization", `Bearer ${process.env.TEST_TOKEN}`)
    .send({
      query: `
        query Component($id: ID!) {
          component(id: $id) {
            id
          }
        }
      `,
      variables: {
        id: "component-id",
      },
    })
    .set("Accept", "application/json");

  expect(JSON.parse(spec.text)).toEqual({
    data: {
      component: null,
    },
  });
});
