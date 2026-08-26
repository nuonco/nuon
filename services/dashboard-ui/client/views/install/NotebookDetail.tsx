import { useParams } from 'react-router'
import {
  keepPreviousData,
  useMutation,
  useQuery,
  useQueryClient,
} from '@tanstack/react-query'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Skeleton } from '@/components/common/Skeleton'
import { NotebookCellCard } from '@/components/notebooks/NotebookCellCard'
import { DetailHeader } from '@/components/layout/DetailHeader'
import { DetailPage } from '@/components/layout/DetailPage'
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
    placeholderData: keepPreviousData,
    queryKey: ['notebook', org?.id, install?.id, notebookId],
    queryFn: () =>
      getNotebook({
        orgId: org!.id,
        installId: install!.id,
        notebookId: notebookId!,
      }),
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
    <>
      <PageTitle segments={[notebook?.name ?? 'Notebook', install?.name]} />
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

      <DetailPage
        header={
          <DetailHeader
            title={notebook?.name || 'Untitled notebook'}
            loading={isLoading}
            loadingWidth={20}
            description={notebook?.description}
            id={notebookId}
          />
        }
      >
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
      </DetailPage>
    </>
  )
}
