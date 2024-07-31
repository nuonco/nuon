import { headers } from "next/headers";
import { getOrg, getInstaller } from "@/app/actions";
import { Link, Video, Card } from "@/components";

export default async function Home({ searchParams }) {
  // Assumption: there is only ever ONE installer
  // TODO: if we move to multiple installers, we will HAVE to use the installer ID in the path
  //       unless we expose a name/slug. this fundamentally changes things. A good interim
  //       solution would be to grab all the installers and return the first.
  // TODO: we can load everything dynamically if we can request the org by slug w/out requiring
  //       the X-Nuon-Org-Id in the request headers when fetching the org itself.
  const headerList = headers();
  const orgId = headerList.get("X-Nuon-Org-Id");
  const { metadata, apps } = await getInstaller(orgId);
  const queryString = new URLSearchParams(searchParams).toString();
  const demoUrl = metadata.formatted_demo_url || metadata.demo_url;
  const isDemoUrlValid = /^((http|https):\/\/)/.test(demoUrl);

  return (
    <>
      <header className="flex flex-auto flex-col gap-6 md:gap-12">
        <div className="text-center">
          <h1>
            <Link href={metadata.homepage_url}>
              <img
                className="inline-block max-w-xl max-h-xl"
                src={metadata.logo_url}
                alt={metadata.name}
              />
            </Link>
          </h1>
        </div>

        <p className="text-xl md:text-4xl text-center leading-relaxed">
          {metadata.description}
        </p>

        {demoUrl && isDemoUrlValid ? <Video src={demoUrl} /> : null}
      </header>

      <main
        className="flex flex-col gap-6"
        data-org-name={headerList.get("X-Nuon-Org-Name")}
        data-org-id={headerList.get("X-Nuon-Org-Id")}
      >
        <div className="grid grid-cols-1 md:grid-cols-2 2xl:grid-cols-4 gap-6">
          {apps.length &&
            apps.map((app) => (
              <Card
                className="p-6 shadow-card-shadow dark:shadow-card-shadow-dark"
                key={app.id}
              >
                <span>
                  <h2 className="text-lg font-semibold mb-2">
                    {app.display_name ? app.display_name : app.name}
                  </h2>
                  <p className="text-xs leading-relaxed">{app.description}</p>
                </span>

                <Link className="text-sm" href={`/${app.name}?${queryString}`}>
                  Install now
                </Link>
              </Card>
            ))}
        </div>
      </main>
    </>
  );
}
