import { AddImageRepository as AddRepository } from '@/components/image-repositories/AddImageRepository'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useOrg } from '@/hooks/use-org'

export const AddImageRepository = () => {
  const { org } = useOrg()
  return (
    <>
      <PageTitle title="Add image repository" />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name },
          { path: `/${org?.id}/settings`, text: 'Settings' },
          {
            path: `/${org?.id}/settings/image-repositories`,
            text: 'Image repositories',
          },
          { path: '', text: 'Add' },
        ]}
      />
      <AddRepository />
    </>
  )
}
