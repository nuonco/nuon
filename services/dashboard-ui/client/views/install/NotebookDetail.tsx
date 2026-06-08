import { useParams } from 'react-router'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { BackLink } from '@/components/common/BackLink'
import { Button } from '@/components/common/Button'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Icon } from '@/components/common/Icon'
import { ID } from '@/components/common/ID'
import { Skeleton } from '@/components/common/Skeleton'
import { Text } from '@/components/common/Text'
import { NotebookCellCard } from '@/components/notebooks/NotebookCellCard'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { createCell, getNotebook } from '@/lib'

export const NotebookDetail = () => {
  const { notebookId } = useParams()
  const { org } = useOrg()
  const { install } = useInstall()
  const queryClient = useQueryClient()

  const { data: notebook, isLoading } = useQuery({
    queryKey: ['notebook', org?.id, install?.id, notebookId],
    queryFn: () =>
      getNotebook({ orgId: org!.id, installId: install!.id, notebookId: notebookId! }),
    enabled: !!org?.id && !!install?.id && !!notebookId,
  })

  const { mutate: addCell, isPending: isAddingCell } = useMutation({
    mutationFn: () =>
      createCell({
        orgId: org!.id,
        installId: install!.id,
        notebookId: notebookId!,
        body: { name: '', inline_contents: '#!/bin/bash\necho hello' },
      }),
    onSuccess: () =>
      queryClient.invalidateQueries({
        queryKey: ['notebook', org?.id, install?.id, notebookId],
      }),
  })

  const cells = notebook?.cells ?? []

  return (
    <PageSection>
      <PageTitle title={`${notebook?.name ?? 'Notebook'} | ${install?.name}`} />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name },
          { path: `/${org?.id}/installs`, text: 'Installs' },
          { path: `/${org?.id}/installs/${install?.id}`, text: install?.name },
          {
            path: `/${org?.id}/installs/${install?.id}/notebooks`,
            text: 'Notebooks',
          },
          {
            path: `/${org?.id}/installs/${install?.id}/notebooks/${notebookId}`,
            text: notebook?.name,
          },
        ]}
      />

      <HeadingGroup>
        <BackLink className="mb-2" />
        <Text variant="h3" weight="strong">
          {notebook?.name || 'Untitled notebook'}
        </Text>
        {notebookId ? <ID>{notebookId}</ID> : null}
      </HeadingGroup>

      {isLoading ? (
        <div className="flex flex-col gap-3">
          <Skeleton height="200px" width="100%" />
          <Skeleton height="200px" width="100%" />
        </div>
      ) : (
        <div className="flex flex-col gap-4">
          {cells.map((cell, i) => (
            <NotebookCellCard
              key={cell.id}
              cell={cell}
              notebookId={notebookId!}
              index={i}
            />
          ))}

          <div>
            <Button
              variant="secondary"
              disabled={isAddingCell}
              onClick={() => addCell()}
            >
              <Icon variant="PlusIcon" size={16} />
              {isAddingCell ? 'Adding cell...' : 'Add cell'}
            </Button>
          </div>
        </div>
      )}
    </PageSection>
  )
}
