import type { Install, InstallSettings } from "../../types";
import { getNodeFields } from "../../utils";

const EAWSRegion = {
  0: "UNSPECIFIED",
  1: "US_EAST_1",
  2: "US_WEST_1",
  3: "US_EAST_2",
  4: "US_WEST_2",
};

export function getInstallSettings(install): InstallSettings {
  let settings = null;
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

export function formatInstall(install): Install {
  const settings = getInstallSettings(install);

  delete install?.gcpSettings;
  delete install?.awsSettings;

  return {
    ...getNodeFields<Install>(install),
    settings,
  };
}
