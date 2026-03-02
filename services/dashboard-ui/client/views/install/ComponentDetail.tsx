import { useParams, useSearchParams } from 'react-router'
import { useQuery } from '@tanstack/react-query'
import { BackLink } from '@/components/common/BackLink'
import { BackToTop } from '@/components/common/BackToTop'
import { ID } from '@/components/common/ID'
import { Text } from '@/components/common/Text'
import { ComponentType } from '@/components/components/ComponentType'
import { InstallComponentBuildCard } from '@/components/install-components/InstallComponentBuildCard'
import {
  InstallComponentConfigCard,
  InstallComponentConfigCardSkeleton,
} from '@/components/install-components/InstallComponentConfigCard'
import { InstallComponentDependencies } from '@/components/install-components/InstallComponentDependencies'
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

  const { data: installComponent, isLoading, error } = useQuery({
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

  console.log(installComponent, componentId)

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
      </div>

      <div className="grid grid-cols-1 md:grid-cols-12 flex-auto divide-x">
        <PageSection className="md:col-span-8">
          {isLoading ? (
            <InstallComponentConfigCardSkeleton />
          ) : installComponent ? (
            <InstallComponentConfigCard config={installComponent?.component?.configs?.[0]} />
          ) : null}

          {component?.dependencies?.length ? (
            <InstallComponentDependencies installComponent={installComponent} />
          ) : null}
        </PageSection>

        <PageSection className="md:col-span-4">
          <Text variant="base" weight="strong">
            Deploy history
          </Text>
          {installComponent ? (
            <InstallComponentBuildCard installComponent={installComponent} shouldPoll />
          ) : null}
        </PageSection>
      </div>

      <BackToTop containerId={CONTAINER_ID} />
    </PageSection>
  )
}
