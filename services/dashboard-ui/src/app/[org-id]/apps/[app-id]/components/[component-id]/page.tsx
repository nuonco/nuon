// NOTE(nnnnat): needs refactored to stratus

import type { Metadata } from 'next'
import { notFound } from 'next/navigation'
import { BackLink } from '@/components/common/BackLink'
import { BackToTop } from '@/components/common/BackToTop'
import { AsyncBoundary } from '@/components/common/AsyncBoundary'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { ID } from '@/components/common/ID'
import { Text } from '@/components/common/Text'
import { ComponentType } from '@/components/components/ComponentType'
import { BuildComponentButton } from '@/components/components/management/BuildComponent'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { getApp, getComponent, getOrg } from '@/lib'
import { Builds, BuildsSkeleton, BuildsError } from './builds'
import { Config, ConfigError, ConfigSkeleton } from './config'

// NOTE: old layout stuff
import { ErrorFallback, Loading, Section } from '@/components'
import { Dependencies } from './dependencies'

export async function generateMetadata({ params }): Promise<Metadata> {
  const {
    ['org-id']: orgId,
    ['app-id']: appId,
    ['component-id']: componentId,
  } = await params
  const [{ data: app }, { data: component }] = await Promise.all([
    getApp({ appId, orgId }),
    getComponent({ componentId, orgId }),
  ])

  return {
    title: `${component?.name} | ${app?.name} | Nuon`,
  }
}

export default async function AppComponent({ params, searchParams }) {
  const {
    ['org-id']: orgId,
    ['app-id']: appId,
    ['component-id']: componentId,
  } = await params
  const sp = await searchParams
  const [{ data: app }, { data: component, error, status }, { data: org }] =
    await Promise.all([
      getApp({ appId, orgId }),
      getComponent({ componentId, orgId }),
      getOrg({ orgId }),
    ])

  if (error) {
    if (status === 404) {
      notFound()
    } else {
      notFound()
    }
  }

  const containerId = 'app-component-page'
  return (
    <PageSection id={containerId} isScrollable className="!p-0 !gap-0">
      <Breadcrumbs
        breadcrumbs={[
          {
            path: `/${orgId}`,
            text: org?.name,
          },
          {
            path: `/${orgId}/apps`,
            text: 'Apps',
          },
          {
            path: `/${orgId}/apps/${appId}`,
            text: app?.name,
          },
          {
            path: `/${orgId}/apps/${appId}/components`,
            text: 'Components',
          },
          {
            path: `/${orgId}/apps/${appId}/components/${componentId}`,
            text: component?.name,
          },
        ]}
      />
      {/* old page layout */}
      <div className="p-6 border-b flex justify-between">
        <HeadingGroup>
          <BackLink className="mb-6" />
          <span className="flex items-center gap-2">
            <ComponentType type={component?.type} displayVariant="icon-only" />
            <Text variant="base" weight="strong">
              {component?.name}
            </Text>
          </span>
          <ID>{component.id}</ID>
        </HeadingGroup>

        <div>
          <BuildComponentButton component={component} variant="primary" />
        </div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-12 flex-auto divide-x">
        <div className="divide-y flex flex-col md:col-span-8">
          {component?.dependencies && (
            <Section className="flex-initial" heading="Dependencies">
              <AsyncBoundary
                errorFallback={ErrorFallback}
                loadingFallback={
                  <Loading
                    variant="stack"
                    loadingText="Loading component dependencies..."
                  />
                }
              >
                <Dependencies component={component} orgId={orgId} />
              </AsyncBoundary>
            </Section>
          )}

          <PageSection>
            <AsyncBoundary
              errorFallback={<ConfigError />}
              loadingFallback={<ConfigSkeleton />}
            >
              <Config componentId={componentId} orgId={orgId} />
            </AsyncBoundary>
          </PageSection>
        </div>
        <div className="divide-y flex flex-col md:col-span-4">
          <PageSection>
            <AsyncBoundary
              errorFallback={<BuildsError />}
              loadingFallback={<BuildsSkeleton />}
            >
              <Builds
                component={component}
                orgId={orgId}
                offset={sp['offset'] || '0'}
              />
            </AsyncBoundary>
          </PageSection>
        </div>
      </div>
      {/* old page layout */}
      <BackToTop containerId={containerId} />
    </PageSection>
  )
}
