import "dotenv/config";
import supertest from "supertest";
import { initServer } from "../../src/server";

const request = supertest(
  initServer({
    org: {
      deleteOrg: jest
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

test("deleteOrg mutation should return false when nothing is deleted", async () => {
  const spec = await request
    .post("/graphql")
    .set("Authorization", `Bearer ${process.env.TEST_TOKEN}`)
    .send({
      query: `
        mutation DeleteOrg($id: ID!) {
          deleteOrg(id: $id)
        }
      `,
    })
    .send({ variables: { id: "org-id" } })
    .set("Accept", "application/json");

  expect(JSON.parse(spec.text)).toEqual({
    data: {
      deleteOrg: false,
    },
  });
});

test("deleteOrg mutation should return true when org is deleted", async () => {
  const spec = await request
    .post("/graphql")
    .set("Authorization", `Bearer ${process.env.TEST_TOKEN}`)
    .send({
      query: `
        mutation DeleteOrg($id: ID!) {
          deleteOrg(id: $id)
        }
      `,
    })
    .send({ variables: { id: "org-id" } })
    .set("Accept", "application/json");

  expect(JSON.parse(spec.text)).toEqual({
    data: {
      deleteOrg: true,
    },
  });
});
