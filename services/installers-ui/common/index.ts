const API_URL = process.env.NUON_API_URL;

//
// pre-existing
//

export async function getInstaller(): Promise<Record<string, any>> {
  const res = await fetch(
    `${API_URL}/v1/installers/${process.env?.NUON_INSTALLER_ID}`,
    {
      cache: "no-store",
      // headers: {
      //   Authorization: `Bearer ${process?.env?.NUON_API_TOKEN}`,
      //   "X-Nuon-Org-ID": orgId,
      // },
    },
  );
  return res.json();
}

export async function getAppBySlug(
  slug: string,
  orgId: string,
): Promise<Record<string, any>> {
  const res = await fetch(`${API_URL}/v1/apps/${slug}`, {
    cache: "no-store",
    headers: {
      Authorization: `Bearer ${process?.env?.NUON_API_TOKEN}`,
      "X-Nuon-Org-ID": orgId,
    },
  });

  return res.json();
}

export async function getInstall(
  id: string,
  orgId: string,
): Promise<Record<string, any>> {
  const res = await fetch(`${API_URL}/v1/installs/${id}`, {
    cache: "no-store",
    headers: {
      Authorization: `Bearer ${process?.env?.NUON_API_TOKEN}`,
      "X-Nuon-Org-ID": orgId,
    },
  });

  return res.json();
}

export async function getCloudPlatformRegions(
  platform: string,
  orgId: string,
): Promise<Array<Record<string, any>>> {
  const res = await fetch(
    `${API_URL}/v1/general/cloud-platform/${platform}/regions`,
    {
      cache: "no-store",
      headers: {
        Authorization: `Bearer ${process?.env?.NUON_API_TOKEN}`,
        "X-Nuon-Org-ID": orgId || "",
      },
    },
  );

  return res.json();
}

export function getFlagEmoji(countryCode = "us") {
  const codePoints = countryCode
    .toUpperCase()
    .split("")
    .map((char) => 127397 + char.charCodeAt(0));
  return String.fromCodePoint(...codePoints);
}
