import { useParams } from 'react-router'
import { ImageRepository as RepositoryDetail } from '@/components/image-repositories/ImageRepository'
import { useImageRepository } from '@/components/image-repositories/queries'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useOrg } from '@/hooks/use-org'

export const ImageRepository = () => {
  const { repositoryId } = useParams()
  const { org } = useOrg()
  const query = useImageRepository(org?.id, repositoryId)
  const title = query.data?.image_url_prefix || 'Image repository'
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
          { path: '', text: title },
        ]}
      />
      <RepositoryDetail key={repositoryId} repositoryId={repositoryId!} />
    </>
  )
}
