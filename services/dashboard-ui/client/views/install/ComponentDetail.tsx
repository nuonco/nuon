import { useParams } from 'react-router'
import { useQuery } from '@tanstack/react-query'
import { BackLink } from '@/components/common/BackLink'
import { BackToTop } from '@/components/common/BackToTop'
import { ID } from '@/components/common/ID'
import { Text } from '@/components/common/Text'
import { ComponentType } from '@/components/components/ComponentType'
import { DeployTimeline } from '@/components/deploys/DeployTimeline'
import {
  InstallComponentConfigCard,
  InstallComponentConfigCardSkeleton,
} from '@/components/install-components/InstallComponentConfigCard'
import { InstallComponentDependencies } from '@/components/install-components/InstallComponentDependencies'
import { ManagementDropdown } from '@/components/install-components/management/ManagementDropdown'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { getInstallComponent } from '@/lib'

const CONTAINER_ID = 'install-component-detail-page'

export const InstallComponentDetail = () => {
  const { componentId } = useParams()
  const { org } = useOrg()
  const { install } = useInstall()

  const { data: installComponent, isLoading } = useQuery({
    queryKey: ['install-component', org?.id, install?.id, componentId],
    queryFn: () =>
      getInstallComponent({
        orgId: org.id,
        installId: install.id,
        componentId: componentId!,
      }),
    enabled: !!org?.id && !!install?.id && !!componentId,
  })

  const component = installComponent?.component
  const latestDeploy = installComponent?.install_deploys?.[0]

  return (
    <PageSection id={CONTAINER_ID} isScrollable className="!p-0 !gap-0">
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name },
          { path: `/${org?.id}/installs`, text: 'Installs' },
          { path: `/${org?.id}/installs/${install?.id}`, text: install?.name },
          {
            path: `/${org?.id}/installs/${install?.id}/components`,
            text: 'Components',
          },
          {
            path: `/${org?.id}/installs/${install?.id}/components/${componentId}`,
            text: component?.name,
          },
        ]}
      />

      <div className="p-6 border-b flex justify-between">
        <HeadingGroup>
          <BackLink className="mb-6" />
          <span className="flex items-center gap-2">
            <ComponentType type={component?.type} displayVariant="icon-only" />
            <Text variant="base" weight="strong">
              {component?.name}
            </Text>
          </span>
          {component?.id ? <ID>{component.id}</ID> : null}
        </HeadingGroup>

        {component && (
          <div className="flex items-center gap-4">
            <ManagementDropdown
              component={component}
              currentBuildId={latestDeploy?.build_id}
              currentDeployStatus={latestDeploy?.status_v2?.status}
              installComponent={installComponent}
            />
          </div>
        )}
      </div>

      <div className="grid grid-cols-1 md:grid-cols-12 flex-auto divide-x">
        <PageSection className="md:col-span-8">
          <>Comonponent config details here!!</>

          {component?.dependencies?.length ? (
            <InstallComponentDependencies deps={component.dependencies} />
          ) : null}
        </PageSection>

        <PageSection className="md:col-span-4">
          <Text variant="base" weight="strong">
            Deploy history
          </Text>
          {component ? (
            <DeployTimeline
              componentId={componentId!}
              componentName={component.name}
              shouldPoll
            />
          ) : null}
        </PageSection>
      </div>

      <BackToTop containerId={CONTAINER_ID} />
    </PageSection>
  )
}
