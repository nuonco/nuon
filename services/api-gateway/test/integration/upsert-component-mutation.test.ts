import "dotenv/config";
import supertest from "supertest";
import { initServer } from "../../src/server";

const request = supertest(initServer());

test("upsertComponent mutation should return null", async () => {
  const spec = await request
    .post("/graphql")
    .set("Authorization", `Bearer ${process.env.TEST_TOKEN}`)
    .send({
      query: `
        mutation UpsertComponent($input: ComponentInput!) {
          upsertComponent(input: $input) {
            id
            name
          }
        }
      `,
    })
    .send({
      variables: {
        input: {
          appId: "app-id",
          name: "Test component",
        },
      },
    })
    .set("Accept", "application/json");

  expect(JSON.parse(spec.text)).toEqual({
    data: {
      upsertComponent: null,
    },
  });
});
