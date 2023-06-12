import "dotenv/config";
import supertest from "supertest";
import { initServer } from "../../src/server";

const request = supertest(initServer({
  component: {
    deleteComponent: jest.fn().mockImplementation((req, cb) => {
      cb(undefined, {
        toObject: jest.fn().mockReturnValue({ deleted: false })
      })
    })
  }
}));

test("deleteComponent mutation should return false when nothing is deleted", async () => {
  const spec = await request
    .post("/graphql")
    .set("Authorization", `Bearer ${process.env.TEST_TOKEN}`)
    .send({
      query: `
        mutation DeleteComponent($id: ID!) {
          deleteComponent(id: $id)
        }
      `,
    })
    .send({ variables: { id: "component-id" } })
    .set("Accept", "application/json");

  expect(JSON.parse(spec.text)).toEqual({
    data: {
      deleteComponent: false,
    },
  });
});
