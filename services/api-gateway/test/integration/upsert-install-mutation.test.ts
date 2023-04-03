import "dotenv/config";
import supertest from "supertest";
import { initServer } from "../../src/server";

const request = supertest(initServer());

test("upsertInstall mutation should return null", async () => {
  const spec = await request
    .post("/graphql")
    .set("Authorization", `Bearer ${process.env.TEST_TOKEN}`)
    .send({
      query: `
        mutation UpsertInstall($input: InstallInput!) {
          upsertInstall(input: $input) {
            id
            name
          }
        }
      `,
    })
    .send({ variables: { input: { appId: "app-id", name: "Test install" } } })
    .set("Accept", "application/json");

  expect(JSON.parse(spec.text)).toEqual({
    data: {
      upsertInstall: null,
    },
  });
});
