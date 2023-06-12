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

const mockInstall = {
  awsSettings: {
    region: 1,
    role: "test:role",
  },
  createdAt: mockDateTimeObject,
  id: "test-id",
  name: "test-node",
  updatedAt: mockDateTimeObject,
};

const request = supertest(initServer({
  install: {
    upsertInstall: jest.fn().mockImplementation((req, cb) =>{
      cb(undefined, {
        toObject: jest.fn().mockReturnValue({ install: mockInstall })
      })
    })
  }
}));

test("upsertInstall mutation should return a new install", async () => {
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
    .send({ variables: { input: { appId: "app-id", name: "test-node" } } })
    .set("Accept", "application/json");

  expect(JSON.parse(spec.text)).toEqual({
    data: {
      upsertInstall: {
        id: "test-id",
        name: "test-node",
      },
    },
  });
});
