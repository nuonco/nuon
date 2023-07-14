import { parseConfigInput, upsertComponent } from "./upsert";

const mockDateTimeObject = {
  day: 31,
  hours: 8,
  minutes: 15,
  month: 12,
  seconds: 30,
  year: 1999,
};

const mockComponent = {
  appId: "app-id",
  createdAt: mockDateTimeObject,
  id: "test-id",
  name: "test-node",
  updatedAt: mockDateTimeObject,
};

const mockComponentServiceClient = {
  upsertComponent: jest
    .fn()
    .mockImplementationOnce((req, cb) => {
      cb(undefined, {
        toObject: jest.fn().mockReturnValue({
          component: mockComponent,
        }),
      });
    })
    .mockImplementationOnce((req, cb) => {
      const err = new Error("some error");
      cb(err);
    }),
};

const mockClients = {
  component: mockComponentServiceClient,
};

test("parseConfigInput should return a component config", () => {
  const spec = parseConfigInput({
    buildConfig: {
      externalImageConfig: {
        ociImageUrl: "test.io/test/iagme",
      },
    },
    deployConfig: {
      basicDeployConfig: {
        healthCheckPath: "/test",
        instanceCount: 2,
        port: 8080,
      },
    },
  });

  expect(spec.toObject()).toMatchInlineSnapshot(`
    {
      "buildCfg": {
        "dockerCfg": undefined,
        "externalImageCfg": {
          "authCfg": {
            "awsIamAuthCfg": undefined,
            "publicAuthCfg": {},
          },
          "ociImageUrl": "test.io/test/iagme",
          "tag": "",
        },
        "helmChartCfg": undefined,
        "noop": undefined,
        "terraformModuleCfg": undefined,
        "timeout": undefined,
      },
      "connections": undefined,
      "deployCfg": {
        "basic": {
          "argsList": [],
          "cpuLimit": "",
          "cpuRequest": "",
          "envVars": undefined,
          "instanceCount": 2,
          "listenerCfg": {
            "healthCheckPath": "/test",
            "listenPort": 8080,
          },
          "memLimit": "",
          "memRequest": "",
        },
        "helmChart": undefined,
        "helmRepo": undefined,
        "noop": undefined,
        "terraformModuleConfig": undefined,
        "timeout": undefined,
      },
      "id": "",
    }
  `);
});

test("upsertComponent resolver should return a basic component", async () => {
  await expect(
    upsertComponent(
      undefined,
      {
        input: {
          appId: "app-id",
          config: { buildConfig: { noop: true }, deployConfig: { noop: true } },
          id: "test-id",
          name: "test-node",
        },
      },
      { clients: mockClients }
    )
  ).resolves.toEqual({
    appId: "app-id",
    config: {
      __typename: "ComponentConfig",
      buildConfig: null,
      deployConfig: null,
    },
    createdAt: "1999-12-31T08:15:30.000Z",
    id: "test-id",
    name: "test-node",
    updatedAt: "1999-12-31T08:15:30.000Z",
  });
});

test("upsertComponent resolver should return error on failed mutation", async () => {
  await expect(
    upsertComponent(
      undefined,
      { input: { appId: "error", id: "error", name: "error" } },
      { clients: mockClients, user: { id: "user-id" } }
    )
  ).rejects.toThrowErrorMatchingInlineSnapshot(`"some error"`);
});

test("upsertComponent resolver should return error if service client doesn't exist", async () => {
  await expect(
    upsertComponent(undefined, { input: {} }, { clients: {} })
  ).rejects.toThrowErrorMatchingInlineSnapshot(`"Service isn't available"`);
});
