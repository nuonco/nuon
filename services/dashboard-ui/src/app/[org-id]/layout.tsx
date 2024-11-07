// @ts-nocheck
import { withPageAuthRequired } from '@auth0/nextjs-auth0'
import { ArrowLineLeft } from '@phosphor-icons/react/dist/ssr'
import { headers } from 'next/headers'
import { notFound } from 'next/navigation'
import { Button, Logo, OrgSwitcher, SignOutButton, MainNav } from '@/components'
import { getOrg, getOrgs } from '@/lib'

export default withPageAuthRequired(
  async function OrgLayout({ children, params }) {
    const orgId = params?.['org-id'] as string
    const [org, orgs] = await Promise.all([
      getOrg({ orgId }).catch((error) => {
        console.error(error)
        notFound()
      }),
      getOrgs().catch((error) => {
        console.error(error)
        notFound()
      }),
    ])

    return (
      <div className="flex min-h-screen">
        <aside className="dashboard_sidebar flex flex-col w-full md:max-w-72">
          <header className="flex flex-col gap-4">
            <div className="border-b flex items-center justify-between px-4 pt-6 pb-4 h-[75px]">
              <Logo />
              {/* <Button className="p-1.5" hasCustomPadding variant="ghost">
                  <ArrowLineLeft />
                  </Button> */}
            </div>

            <div className="px-4">
              <OrgSwitcher initOrg={org} initOrgs={orgs} />
            </div>
          </header>

          <div className="dashboard_nav flex-auto flex flex-col justify-between px-4 pb-6 pt-8">
            <MainNav orgId={orgId} />

            <div>
              <SignOutButton />
            </div>
          </div>
        </aside>
        <div className="dashboard_content h-screen flex-auto md:border-l">
          {children}
        </div>
      </div>
    )
  },
  {
    returnTo() {
      return headers().get('x-origin-url')
    },
  }
)
