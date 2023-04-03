import { TInstall } from "../../types";
import { getNodeFields } from "../../utils";

const EAWSRegion = {
  0: "UNSPECIFIED",
  1: "US_EAST_1",
  2: "US_WEST_1",
  3: "US_EAST_2",
  4: "US_WEST_2",
};

export function getInstallSettings(install) {
  let settings = { __typename: "NoopSettings" };
  const { awsSettings, gcpSettings } = install;

  if (gcpSettings) {
    settings = {
      __typename: "GCPSettings",
      ...gcpSettings,
    };
  } else {
    settings = {
      ...awsSettings,
      __typename: "AWSSettings",
      region: EAWSRegion[awsSettings.region],
    };
  }

  return settings;
}

export function formatInstall(install): TInstall {
  const settings = getInstallSettings(install);

  delete install?.gcpSettings;
  delete install?.awsSettings;

  return {
    ...getNodeFields<TInstall>(install),
    settings,
  };
}
