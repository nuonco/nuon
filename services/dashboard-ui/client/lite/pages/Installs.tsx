import { Button } from '../components/atoms/Button'
import { InstallsTable } from '../components/organisms/InstallsTable'
import { useBreadcrumbs } from '../hooks/use-breadcrumbs'
import { usePageTitle } from '../hooks/use-page-title'
import { installSetupHref } from '../utils/hrefs'
import { useOrg } from '../providers/org-provider'

export const Installs = () => {
  const { org, orgId } = useOrg()

  usePageTitle('Installs')
  useBreadcrumbs([
    {
      label: org?.name,
      href: orgId ? `/${orgId}` : undefined,
      loadingWidth: 16,
    },
    { label: 'Installs' },
  ])

  return (
    <div className="flex w-full flex-col gap-6">
      <div className="flex justify-end">
        <Button
          variant="primary"
          href={orgId ? installSetupHref(orgId) : undefined}
          disabled={!orgId}
        >
          Create install
        </Button>
      </div>
      <InstallsTable />
    </div>
  )
}
