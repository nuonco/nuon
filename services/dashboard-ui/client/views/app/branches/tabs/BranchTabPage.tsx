import type { ReactNode } from 'react'
import { useParams } from 'react-router'
import { PageSection } from '@/components/layout/PageSection'
import { SectionHeader } from '@/components/layout/SectionHeader'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useApp } from '@/hooks/use-app'
import { useBranch } from '@/hooks/use-branch'
import { useOrg } from '@/hooks/use-org'

interface IBranchTabPage {
  tab: string
  heading: string
  subheading?: string
  actions?: ReactNode
  children: ReactNode
}

export const BranchTabPage = ({
  tab,
  heading,
  subheading,
  actions,
  children,
}: IBranchTabPage) => {
  const { org } = useOrg()
  const { app } = useApp()
  const { branch } = useBranch()
  const params = useParams()
  const basePath = `/${org?.id}/apps/${app?.id}/branches/${params.branchId}`

  return (
    <PageSection>
      <PageTitle
        segments={[branch?.name ? `${branch.name} ${tab.toLowerCase()}` : tab, app?.name]}
      />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name },
          { path: `/${org?.id}/apps`, text: 'Apps' },
          { path: `/${org?.id}/apps/${app?.id}`, text: app?.name },
          { path: basePath, text: branch?.name },
          { path: `${basePath}/${tab.toLowerCase()}`, text: tab },
        ]}
      />
      <SectionHeader
        title={heading}
        description={subheading}
        actions={actions}
      />
      {children}
    </PageSection>
  )
}
