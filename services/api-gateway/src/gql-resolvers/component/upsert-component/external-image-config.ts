import {
  AWSIAMAuthCfg,
  ExternalImageAuthConfig,
  ExternalImageConfig,
  PublicAuthCfg,
} from "../../../build/components/build/v1/external_image_pb";
import type { ExternalImageInput, TgRPCMessage } from "../../../types";

export function initExternalImageConfig(
  externalImageInput: ExternalImageInput
): TgRPCMessage {
  const { authConfig, ociImageUrl, tag } = externalImageInput;

  const externalImageAuthCfg = new ExternalImageAuthConfig();
  if (authConfig) {
    const privateAuthCfg = new AWSIAMAuthCfg()
      .setIamRoleArn(authConfig.role)
      .setAwsRegion(authConfig.region);
    externalImageAuthCfg.setAwsIamAuthCfg(privateAuthCfg);
  } else {
    const publicAuthCfg = new PublicAuthCfg();
    externalImageAuthCfg.setPublicAuthCfg(publicAuthCfg);
  }

  return new ExternalImageConfig()
    .setOciImageUrl(ociImageUrl)
    .setTag(tag)
    .setAuthCfg(externalImageAuthCfg);
}
