import { PolicyAnalytics as PolicyAnalyticsComponent } from '@/components/policies/PolicyAnalytics'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useApp } from '@/hooks/use-app'

export const PolicyAnalytics = () => {
  const { app } = useApp()
  return (
    <>
      <PageTitle segments={['Policy analytics', app?.name]} />
      <PolicyAnalyticsComponent />
    </>
  )
}
