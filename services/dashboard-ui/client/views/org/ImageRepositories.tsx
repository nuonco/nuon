import { ImageRepositories as RepositoryList } from '@/components/image-repositories/ImageRepositories'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useOrg } from '@/hooks/use-org'

export const ImageRepositories = () => {
  const { org } = useOrg()
  return (
    <>
      <PageTitle title="Image repositories" />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name },
          { path: `/${org?.id}/settings`, text: 'Settings' },
          {
            path: `/${org?.id}/settings/image-repositories`,
            text: 'Image repositories',
          },
        ]}
      />
      <RepositoryList />
    </>
  )
}
