import "dotenv/config";
import supertest from "supertest";
import { initServer } from "../../src/server";

const request = supertest(
  initServer({
    app: {
      deleteApp: jest
        .fn()
        .mockImplementationOnce((req, cb) => {
          cb(undefined, {
            toObject: jest.fn().mockReturnValue({ deleted: false }),
          });
        })
        .mockImplementationOnce((req, cb) => {
          cb(undefined, {
            toObject: jest.fn().mockReturnValue({ deleted: true }),
          });
        }),
    },
  })
);

test("deleteApp mutation should return false when nothing is deleted", async () => {
  const spec = await request
    .post("/graphql")
    .set("Authorization", `Bearer ${process.env.TEST_TOKEN}`)
    .send({
      query: `
        mutation DeleteApp($id: ID!) {
          deleteApp(id: $id)
        }
      `,
    })
    .send({ variables: { id: "app-id" } })
    .set("Accept", "application/json");

  expect(JSON.parse(spec.text)).toEqual({
    data: {
      deleteApp: false,
    },
  });
});

test("deleteApp mutation should return true when app is deleted", async () => {
  const spec = await request
    .post("/graphql")
    .set("Authorization", `Bearer ${process.env.TEST_TOKEN}`)
    .send({
      query: `
        mutation DeleteApp($id: ID!) {
          deleteApp(id: $id)
        }
      `,
    })
    .send({ variables: { id: "app-id" } })
    .set("Accept", "application/json");

  expect(JSON.parse(spec.text)).toEqual({
    data: {
      deleteApp: true,
    },
  });
});
