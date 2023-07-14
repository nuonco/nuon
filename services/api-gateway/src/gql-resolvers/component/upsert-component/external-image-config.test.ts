import { AwsRegion } from "../../../types";
import { initExternalImageConfig } from "./external-image-config";

test("initExternalImageConfig should return a gRPC message for a public docker hub image", () => {
  const spec = initExternalImageConfig({
    ociImageUrl: "some-place.io/test/image",
    tag: "latest",
  });

  expect(spec.toObject()).toMatchInlineSnapshot(`
    {
      "authCfg": {
        "awsIamAuthCfg": undefined,
        "publicAuthCfg": {},
      },
      "ociImageUrl": "some-place.io/test/image",
      "tag": "latest",
    }
  `);
});

test("initExternalImageConfig should return a gRPC message for a private ECR image", () => {
  const spec = initExternalImageConfig({
    authConfig: {
      region: AwsRegion.UsEast_1,
      role: "test",
    },
    ociImageUrl: "some-place.io/test/image",
    tag: "latest",
  });

  expect(spec.toObject()).toMatchInlineSnapshot(`
    {
      "authCfg": {
        "awsIamAuthCfg": {
          "awsRegion": "US_EAST_1",
          "iamRoleArn": "test",
        },
        "publicAuthCfg": undefined,
      },
      "ociImageUrl": "some-place.io/test/image",
      "tag": "latest",
    }
  `);
});
