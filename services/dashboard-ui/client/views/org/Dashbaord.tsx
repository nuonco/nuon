import { PageLayout } from '@/components/layout/PageLayout'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { useOrg } from '@/hooks/use-org'

export const Dashboard = () => {
  const { org } = useOrg()

  if (!org) return <>Loading org...</>
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
