import "dotenv/config";
import supertest from "supertest";
import { initServer } from "../../src/server";

const request = supertest(initServer());

// TODO: this test will return the user ID of the access token

test.skip("Me query should return null", async () => {
  const spec = await request
    .post("/graphql")
    .set("Authorization", `Bearer ${process.env.TEST_TOKEN}`)
    .send({
      query: "query Me { me { id } }",
    })
    .set("Accept", "application/json");

  expect(JSON.parse(spec.text)).toEqual({
    data: {
      me: {
        id: "BUfg2P1FO9gzEYNang9rjTfqEUNANF0y@clients",
      },
    },
  });
});
