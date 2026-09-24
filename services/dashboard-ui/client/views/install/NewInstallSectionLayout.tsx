import { Outlet } from 'react-router'
import { PageSection } from '@/components/layout/PageSection'
import { SectionHeader } from '@/components/layout/SectionHeader'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { TabNav } from '@/components/navigation/TabNav'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import type { TNavLink } from '@/types'

export const NEW_INSTALL_RESOURCES_TABS: TNavLink[] = [
  { path: '/', text: 'Stack' },
  { path: '/sandbox', text: 'Sandbox' },
  { path: '/components', text: 'Components' },
  { path: '/images', text: 'Images' },
]

export const NEW_INSTALL_OPERATIONS_TABS: TNavLink[] = [
  { path: '/', text: 'Activity' },
  { path: '/actions', text: 'Actions' },
  { path: '/runbooks', text: 'Runbooks' },
  { path: '/policies', text: 'Policies' },
  { path: '/runner', text: 'Runner' },
]

export const NEW_INSTALL_CONFIGURATION_TABS: TNavLink[] = [
  { path: '/', text: 'App branch' },
  { path: '/inputs', text: 'Inputs' },
  { path: '/config-file', text: 'Config file' },
  { path: '/overrides', text: 'Overrides' },
  { path: '/state', text: 'State' },
]

const NewInstallSectionLayout = ({
  path,
  tabs,
  title,
}: {
  path: string
  tabs: TNavLink[]
  title: string
}) => {
  const { org } = useOrg()
  const { install } = useInstall()
  const basePath = `/${org?.id}/installs/${install?.id}${path}`

  return (
    <PageSection>
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name },
          { path: `/${org?.id}/installs`, text: 'Installs' },
          { path: `/${org?.id}/installs/${install?.id}`, text: install?.name },
          { path: basePath, text: title },
        ]}
      />
      <SectionHeader title={title} />
      <TabNav basePath={basePath} tabs={tabs} />
      <Outlet />
    </PageSection>
  )
}

export const NewInstallResourcesLayout = () => (
  <NewInstallSectionLayout
    path="/resources"
    tabs={NEW_INSTALL_RESOURCES_TABS}
    title="Resources"
  />
)

export const NewInstallOperationsLayout = () => (
  <NewInstallSectionLayout
    path="/operations"
    tabs={NEW_INSTALL_OPERATIONS_TABS}
    title="Operations"
  />
)

export const NewInstallConfigurationLayout = () => (
  <NewInstallSectionLayout
    path="/configuration"
    tabs={NEW_INSTALL_CONFIGURATION_TABS}
    title="Configuration"
  />
)
