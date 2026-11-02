import { useParams } from 'react-router'
import { AddImageRepository } from '@/components/image-repositories/AddImageRepository'
import { ImageRepositories as RepositoryList } from '@/components/image-repositories/ImageRepositories'
import { ImageRepository } from '@/components/image-repositories/ImageRepository'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useOrg } from '@/hooks/use-org'

export const ImageRepositories = ({ adding = false }: { adding?: boolean }) => {
  const { org } = useOrg()
  const { repositoryId } = useParams()
  const title = adding
    ? 'Add image repository'
    : repositoryId
      ? 'Image repository'
      : 'Image repositories'
  return (
    <>
      <PageTitle title={title} />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name },
          { path: `/${org?.id}/settings`, text: 'Settings' },
          {
            path: `/${org?.id}/settings/image-repositories`,
            text: 'Image repositories',
          },
          ...(adding || repositoryId
            ? [{ path: '', text: adding ? 'Add' : 'Setup' }]
            : []),
        ]}
      />
      {adding ? (
        <AddImageRepository />
      ) : repositoryId ? (
        <ImageRepository key={repositoryId} repositoryId={repositoryId} />
      ) : (
        <RepositoryList />
      )}
    </>
  )
}
