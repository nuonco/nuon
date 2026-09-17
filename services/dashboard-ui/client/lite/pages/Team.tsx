import { Button } from '../components/atoms/Button'
import { InviteMember } from '../components/organisms/InviteMember'
import { TeamTable } from '../components/organisms/TeamTable'
import { useBreadcrumbs } from '../hooks/use-breadcrumbs'
import { usePageTitle } from '../hooks/use-page-title'
import { useSurfaces } from '../hooks/use-surfaces'
import { useOrg } from '../providers/org-provider'

export const Team = () => {
  const { org, orgId } = useOrg()
  const { openModal } = useSurfaces()

  usePageTitle('Team')
  useBreadcrumbs([
    {
      label: org?.name,
      href: orgId ? `/${orgId}` : undefined,
      loadingWidth: 16,
    },
    { label: 'Team' },
  ])

  return (
    <div className="flex w-full flex-col gap-6">
      <div className="flex justify-end">
        <Button
          variant="primary"
          disabled={!orgId}
          onClick={() => openModal(<InviteMember />)}
        >
          Invite team member
        </Button>
      </div>
      <TeamTable />
    </div>
  )
}
