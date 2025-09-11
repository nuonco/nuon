import { redirect } from 'next/navigation'
import { getOrgIdFromCookie } from "@/actions/orgs/org-session-cookie";
import { Dashboard, Text } from '@/components'
import { auth0 } from '@/lib/auth'
import { getOrgs } from "@/lib"

export default async function Home() {
  const session = await auth0.getSession();

  if (session) {
    const orgId = await getOrgIdFromCookie()
    const { data: orgs } = await getOrgs();

    if (orgs && orgs?.length) {
      if (orgId && orgs.some((org) => org.id === orgId)) {
        redirect(`/${orgId}/apps`);
      } else {
        redirect(`/${orgs?.at(0)?.id}/apps`);
      }
    } else {
      redirect(`/request-access`);
    }
  }
  
  return (
    <Dashboard isLandingPage>
      <main className="flex flex-col h-full gap-24 md:gap-[250px] py-6 md:py-12 lg:py-24">
        <div className="flex flex-col gap-6 lg:max-w-xl">
          <Text
            className="!text-[28px] !leading-[30px] md:!text-[52px] md:!leading-[58px] !inline"
            variant="semi-18"
            level={1}
          >
            <span className="text-gradient inline-flex">
              Bring Your Own Cloud
            </span>, <br />
            for everyone.
          </Text>
          <Text className="!text-lg md:!text-xl !leading-loose">
            If you already have an account, please log in.
            Otherwise you will be directed to request access.
          </Text>
          <div className="flex items-center gap-4 mt-4">
            <a
              className="flex flex-initial items-center w-fit gap-1 bg-primary-600 hover:bg-primary-700 focus:bg-primary-700 active:bg-primary-900 rounded-md text-lg text-cool-grey-50 border border-transparent px-5 py-1.5 font-medium"
              href="/api/auth/login?returnTo=/"
            >
              Login
            </a>
            <a
              className="flex flex-initial items-center w-fit gap-1 bg-white text-cool-grey-950 dark:bg-dark-grey-100 dark:text-cool-grey-50 hover:bg-black/5 focus:bg-black/5 active:bg-black/10 dark:hover:bg-white/5 dark:focus:bg-white/5 dark:active:bg-white/10 rounded-md text-lg border px-3 py-1.5 font-medium"
              href="/api/auth/login?returnTo=/request-access"
            >
              Request access
            </a>
          </div>
        </div>
        <div className="bg-[#1D0B2F] border-primary-950/50 border shadow text-cool-grey-50 rounded-lg p-8 md:p-12 landing-toml">
          <div className="flex flex-col gap-8 max-w-md">
            <Text
              className="!text-[26px] md:!text-[32px] md:!leading-[32px] !inline"
              variant="semi-18"
            >
              Explore our docs
            </Text>
            <Text className="!text-md md:!text-lg !leading-loose" variant="reg-14">
              Designed by developers, for developers. Learn more about how Nuon
              works, configure your first app and build Nuon directly into your
              signup flow and control plane using our API and SDKs.
            </Text>
            <a
              className="flex flex-initial items-center w-fit gap-1 bg-dark-grey-100/20 text-cool-grey-50 hover:bg-white/5 focus:bg-white/5 active:bg-black/10 rounded-md text-lg border border-cool-grey-50/20 px-3 py-1.5 font-medium"
              target="_blank"
              href="https://docs.nuon.co"
            >
              Go to docs
            </a>
          </div>
        </div>
      </main>
    </Dashboard>
  )
}
