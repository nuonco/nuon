import { useNavigate } from 'react-router'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Button } from '@/components/common/Button'
import { EmptyState } from '@/components/common/EmptyState'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Icon } from '@/components/common/Icon'
import { ID } from '@/components/common/ID'
import { Link } from '@/components/common/Link'
import { Skeleton } from '@/components/common/Skeleton'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { createNotebook, getNotebooks } from '@/lib'

export const Notebooks = () => {
  const { org } = useOrg()
  const { install } = useInstall()
  const navigate = useNavigate()
  const queryClient = useQueryClient()

  const { data: notebooks, isLoading } = useQuery({
    queryKey: ['notebooks', org?.id, install?.id],
    queryFn: () => getNotebooks({ orgId: org!.id, installId: install!.id }),
    enabled: !!org?.id && !!install?.id,
  })

  const { mutate: create, isPending: isCreating } = useMutation({
    mutationFn: () =>
      createNotebook({
        orgId: org!.id,
        installId: install!.id,
        body: { name: 'Untitled notebook' },
      }),
    onSuccess: (nb) => {
      queryClient.invalidateQueries({ queryKey: ['notebooks', org?.id, install?.id] })
      navigate(`/${org?.id}/installs/${install?.id}/notebooks/${nb.id}`)
    },
  })

  return (
    <PageSection>
      <PageTitle title={`Notebooks | ${install?.name}`} />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name },
          { path: `/${org?.id}/installs`, text: 'Installs' },
          { path: `/${org?.id}/installs/${install?.id}`, text: install?.name },
          {
            path: `/${org?.id}/installs/${install?.id}/notebooks`,
            text: 'Notebooks',
          },
        ]}
      />
      <div className="flex items-start justify-between gap-4">
        <HeadingGroup>
          <Text variant="base" weight="strong">
            Notebooks
          </Text>
          <Text variant="subtext" theme="neutral">
            Run commands on the runner for this install.
          </Text>
        </HeadingGroup>
        <Button
          variant="primary"
          disabled={isCreating}
          onClick={() => create()}
        >
          <Icon variant="PlusIcon" size={16} />
          {isCreating ? 'Creating...' : 'Create notebook'}
        </Button>
      </div>

      {isLoading ? (
        <div className="flex flex-col gap-3">
          <Skeleton height="64px" width="100%" />
          <Skeleton height="64px" width="100%" />
        </div>
      ) : notebooks?.length ? (
        <div className="flex flex-col divide-y rounded-md border">
          {notebooks.map((nb) => (
            <Link
              key={nb.id}
              href={`/${org?.id}/installs/${install?.id}/notebooks/${nb.id}`}
              className="flex items-center justify-between gap-4 p-4 hover:bg-black/5 dark:hover:bg-white/5"
            >
              <div className="flex flex-col gap-1 min-w-0">
                <Text weight="strong">{nb.name || 'Untitled notebook'}</Text>
                <ID>{nb.id}</ID>
              </div>
              <Time variant="subtext" time={nb.updated_at} format="relative" />
            </Link>
          ))}
        </div>
      ) : (
        <EmptyState
          variant="table"
          emptyTitle="No notebooks yet"
          emptyMessage="Create a notebook to start running commands on this install."
          action={
            <Button variant="primary" disabled={isCreating} onClick={() => create()}>
              <Icon variant="PlusIcon" size={16} />
              {isCreating ? 'Creating...' : 'Create notebook'}
            </Button>
          }
        />
      )}
    </PageSection>
  )
}
