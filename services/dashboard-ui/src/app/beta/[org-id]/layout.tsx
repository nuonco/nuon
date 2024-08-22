import { GoSidebarCollapse } from 'react-icons/go'
import { Button, Logo, OrgSwitcher, SignOutButton, MainNav } from '@/components'
import { getOrg, getOrgs } from '@/lib'

export default async function OrgLayout({ children, params }) {
  const orgId = params?.['org-id'] as string
  const [org, orgs] = await Promise.all([getOrg({ orgId }), getOrgs()])

  return (
    <div className="flex min-h-screen">
      <aside className="dashboard_sidebar flex flex-col w-full md:w-72">
        <header className="flex flex-col gap-4">
          <div className="border-b flex items-center justify-between px-4 pt-6 pb-4">
            <Logo />
            <Button className="px-1.5" variant="ghost">
              <GoSidebarCollapse className="rotate-180" />
            </Button>
          </div>

          <div className="px-4">
            <OrgSwitcher initOrg={org} initOrgs={orgs} />
          </div>
        </header>

        <div className="flex-auto flex flex-col justify-between px-4 pb-6 pt-8">
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
}
