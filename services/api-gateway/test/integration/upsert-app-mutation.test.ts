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

const mockApp = {
  createdAt: mockDateTimeObject,
  id: "test-id",
  name: "test-node",
  updatedAt: mockDateTimeObject,
};

const request = supertest(
  initServer({
    app: {
      upsertApp: jest.fn().mockImplementation((req, cb) => {
        cb(undefined, {
          toObject: jest.fn().mockReturnValue({ app: mockApp }),
        });
      }),
    },
  })
);

test("upsertApp mutation should return a new app", async () => {
  const spec = await request
    .post("/graphql")
    .set("Authorization", `Bearer ${process.env.TEST_TOKEN}`)
    .send({
      query: `
        mutation UpsertApp($input: AppInput!) {
          upsertApp(input: $input) {
            id
            name
          }
        }
      `,
    })
    .send({ variables: { input: { name: "test-node", orgId: "org-id" } } })
    .set("Accept", "application/json");

  expect(JSON.parse(spec.text)).toEqual({
    data: {
      upsertApp: {
        id: "test-id",
        name: "test-node",
      },
    },
  });
});
