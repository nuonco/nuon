import "dotenv/config";
import supertest from "supertest";
import { initServer } from "../../src/server";

const mockDateTimeObject = {
  day: 31,
  hours: 8,
  minutes: 15,
  month: 12,
  seconds: 30,
  year: 1999,
};

const mockComponent = {
  createdAt: mockDateTimeObject,
  id: "test-id",
  name: "test-node",
  updatedAt: mockDateTimeObject,
};

const request = supertest(
  initServer({
    component: {
      getComponent: jest.fn().mockImplementation((req, cb) => {
        cb(undefined, {
          toObject: jest.fn().mockReturnValue({ component: mockComponent }),
        });
      }),
    },
  })
);

test("Component query should return a component", async () => {
  const spec = await request
    .post("/graphql")
    .set("Authorization", `Bearer ${process.env.TEST_TOKEN}`)
    .send({
      query: `
        query Component($id: ID!) {
          component(id: $id) {
            createdAt
            id
            name
            updatedAt
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
      component: {
        createdAt: "1999-12-31T08:15:30.000Z",
        id: "test-id",
        name: "test-node",
        updatedAt: "1999-12-31T08:15:30.000Z",
      },
    },
  });
});
