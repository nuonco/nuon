import "dotenv/config";
import supertest from "supertest";
import { initServer } from "../../src/server";

const request = supertest(initServer());

test("Components query should return total count of 0", async () => {
  const spec = await request
    .post("/graphql")
    .set("Authorization", `Bearer ${process.env.TEST_TOKEN}`)
    .send({
      query: `
        query Components($appId: ID!, $options: ConnectionOptions) {
          components(appId: $appId, options: $options) {
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
      components: {
        totalCount: 0,
      },
    },
  });
});
