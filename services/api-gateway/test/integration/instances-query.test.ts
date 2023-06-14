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

const mockInstance = {
  createdAt: mockDateTimeObject,
  componentId: "component-id",
  id: "test-id",
  updatedAt: mockDateTimeObject,
};

const mockBuild = {
  createdAt: mockDateTimeObject,
  id: "build-id",
  updatedAt: mockDateTimeObject,
};

const mockComponent = {
  createdAt: mockDateTimeObject,
  id: "component-id",
  updatedAt: mockDateTimeObject,
};

const mockDeploy = {
  createdAt: mockDateTimeObject,
  id: "deploy-id",
  updatedAt: mockDateTimeObject,
};

const request = supertest(
  initServer({
    build: {
      listBuildsByInstance: jest.fn().mockImplementation((req, cb) => {
        cb(undefined, {
          toObject: jest.fn().mockReturnValue({ buildsList: [mockBuild] }),
        });
      }),
    },
    component: {
      getComponent: jest.fn().mockImplementation((req, cb) => {
        cb(undefined, {
          toObject: jest.fn().mockReturnValue({ component: mockComponent }),
        });
      }),
    },
    deploy: {
      getDeploysByInstance: jest.fn().mockImplementation((req, cb) => {
        cb(undefined, {
          toObject: jest.fn().mockReturnValue({ deploysList: [mockDeploy] }),
        });
      }),
    },
    instance: {
      getInstancesByInstall: jest.fn().mockImplementation((req, cb) => {
        cb(undefined, {
          toObject: jest
            .fn()
            .mockReturnValue({ instancesList: [mockInstance] }),
        });
      }),
    },
  })
);

test("Instances query should return hydrated instances", async () => {
  const spec = await request
    .post("/graphql")
    .set("Authorization", `Bearer ${process.env.TEST_TOKEN}`)
    .send({
      query: `
        query Instances($installId: ID!) {
          instances(installId: $installId) {
            build {
              id
            }
            component {
              id
            }
            deploy {
              id
            }
            id
          }
        }
      `,
      variables: {
        installId: "install-id",
      },
    })
    .set("Accept", "application/json");

  expect(JSON.parse(spec.text)).toEqual({
    data: {
      instances: [
        {
          build: {
            id: "build-id",
          },
          component: {
            id: "component-id",
          },
          deploy: {
            id: "deploy-id",
          },
          id: "test-id",
        },
      ],
    },
  });
});
