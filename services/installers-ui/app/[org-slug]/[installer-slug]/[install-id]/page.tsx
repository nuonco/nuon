import { headers } from "next/headers";
import { createInstall, getInstall, redeployInstall } from "@/app/actions";
import { getInstaller, getAppBySlug } from "@/app/actions";

import { Link } from "@/components";
import InstallStepper from "@/components/InstallStepper";
import { getCloudPlatformRegions } from "@/common";

export default async function Installer({ params, searchParams }) {
  const headerList = headers();
  const orgId = headerList.get("X-Nuon-Org-Id");
  const slug = params?.["installer-slug"];
  const installId = params?.["install-id"];

  const [app, installer, install] = await Promise.all([
    getAppBySlug(slug, orgId),
    getInstaller(orgId),
    getInstall(installId, orgId),
  ]);
  const regions = await getCloudPlatformRegions(app.cloud_platform, orgId);

  return (
    <>
      <header className="flex flex-col gap-4">
        <div className="flex justify-between items-center pt-4">
          <Link className="w-fit" href="/">
            {"< Other installation options"}
          </Link>
          <Link href={installer?.metadata?.homepage_url}>
            <img
              className="inline-block"
              src={installer?.metadata?.logo_url}
              alt={installer?.metadata?.name}
            />
          </Link>
        </div>

        <div className="text-center">
          <h1 className="text-2xl font-semibold mb-2">
            {app.display_name || app.name}
          </h1>
          <p>{app?.description}</p>
        </div>
      </header>

      <main className="flex-auto" id="steps">
        <InstallStepper
          app={app}
          existingInstall={install}
          installer={installer}
          searchParams={searchParams}
          createInstall={createInstall}
          getInstall={getInstall}
          redeployInstall={redeployInstall}
          regions={regions}
        />
      </main>
    </>
  );
}
