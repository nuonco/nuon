import "dotenv/config";
import supertest from "supertest";
import { initServer } from "../../src/server";

const request = supertest(
  initServer({
    status: {
      ping: jest.fn().mockImplementation((req, cb) => {
        cb(undefined, { getStatus: jest.fn().mockReturnValue("ok") });
      }),
    },
  })
);

test("ping query should return 'ok' when ping service is available", async () => {
  const spec = await request
    .post("/graphql")
    .set("Authorization", `Bearer ${process.env.TEST_TOKEN}`)
    .send({
      query: "query Ping { ping }",
    })
    .set("Accept", "application/json");

  expect(JSON.parse(spec.text)).toEqual({
    data: {
      ping: "ok",
    },
  });
});
