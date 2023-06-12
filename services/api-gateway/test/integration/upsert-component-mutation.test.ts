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
      upsertComponent: jest.fn().mockImplementation((req, cb) => {
        cb(undefined, {
          toObject: jest.fn().mockReturnValue({ component: mockComponent }),
        });
      }),
    },
  })
);

test("upsertComponent mutation should return a new component", async () => {
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
          name: "test-node",
        },
      },
    })
    .set("Accept", "application/json");

  expect(JSON.parse(spec.text)).toEqual({
    data: {
      upsertComponent: {
        id: "test-id",
        name: "test-node",
      },
    },
  });
});
