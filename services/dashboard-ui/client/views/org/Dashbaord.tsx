import { useEffect } from 'react'
import { useNavigate } from 'react-router'
import { PageLayout } from '@/components/layout/PageLayout'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { useOrg } from '@/hooks/use-org'

export const Dashboard = () => {
  const navigate = useNavigate()
  const { org } = useOrg()

  useEffect(() => {
    if (org?.features?.['org-dashboard'] === false) {
      navigate(`/${org.id}/apps`)
    }
  }, [])

  return (
    <PageLayout>
      <Breadcrumbs
        breadcrumbs={[
          {
            path: `/${org.id}`,
            text: org?.name,
          },
        ]}
      />
      <span>Org dashboard</span>
    </PageLayout>
  )
}
