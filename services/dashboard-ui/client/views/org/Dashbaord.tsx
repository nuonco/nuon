import { PageLayout } from '@/components/layout/PageLayout'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { useOrg } from '@/hooks/use-org'

export const Dashboard = () => {
  const { org } = useOrg()

  return (
    <PageLayout>
      <Breadcrumbs
        breadcrumbs={[
          {
            path: `/${org.id}`,
            text: org?.name,
          }          
        ]}
      />
      <span>Org dashboard</span>
    </PageLayout>
  )
}
