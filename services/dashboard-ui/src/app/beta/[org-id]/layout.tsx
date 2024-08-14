import {
  Dropdown,
  Logo,
  Link,
  OrgSwitcher,
  ProfileDropdown,
} from '@/components'
import { getOrg, getOrgs } from '@/lib'

export default async function OrgLayout({ children, params }) {
  const orgId = params?.['org-id'] as string
  const [org, orgs] = await Promise.all([getOrg({ orgId }), getOrgs()])

  return (
    <div className="flex min-h-screen">
      <aside className="flex flex-col w-full md:w-72">
        <header className="flex flex-col px-4 py-6 gap-4">
          <Logo />
          <OrgSwitcher initOrg={org} initOrgs={orgs} />
        </header>

        <div className="flex-auto flex flex-col justify-between px-4 pb-6 pt-8">
          <nav className="flex-auto">
            <Link href={`/beta/${orgId}/apps`}>Apps</Link>
            <Link href={`/beta/${orgId}/installs`}>Installs</Link>
          </nav>

          <div>
            <ProfileDropdown />
          </div>
        </div>
      </aside>
      <div className="flex-auto md:border-l">
        <header className="flex justify-between items-center border-b px-6 py-4">
          <div className="flex-auto">
            <input
              className="rounded bg-transparent border px-2 py-1"
              placeholder="Search"
              type="search"
            />
          </div>
          <div>
            <Link href="https://docs.nuon.co" target="_blank">
              Docs
            </Link>
          </div>
        </header>
        <main className="px-6 py-8 min-h-full flex flex-col gap-8">
          {children}
        </main>
      </div>
    </div>
  )
}
