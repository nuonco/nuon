import "dotenv/config";
import supertest from "supertest";
import { initServer } from "../../src/server";

const request = supertest(initServer());

test("Installs query should return total count of 0", async () => {
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
        totalCount: 0,
      },
    },
  });
});
