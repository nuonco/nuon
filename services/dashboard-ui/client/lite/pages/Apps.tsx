import { Button } from '../components/atoms/Button'
import { AppsTable } from '../components/organisms/AppsTable'
import { useBreadcrumbs } from '../hooks/use-breadcrumbs'
import { usePageTitle } from '../hooks/use-page-title'
import { appSetupHref } from '../utils/hrefs'
import { useOrg } from '../providers/org-provider'

export const Apps = () => {
  const { org, orgId } = useOrg()

  usePageTitle('Apps')
  useBreadcrumbs([
    {
      label: org?.name,
      href: orgId ? `/${orgId}` : undefined,
      loadingWidth: 16,
    },
    { label: 'Apps' },
  ])

  return (
    <div className="flex w-full flex-col gap-6">
      <div className="flex justify-end">
        <Button
          variant="primary"
          href={orgId ? appSetupHref(orgId) : undefined}
          disabled={!orgId}
        >
          Create app
        </Button>
      </div>
      <AppsTable />
    </div>
  )
}
