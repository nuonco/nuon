import type { Metadata } from 'next'
import { notFound } from 'next/navigation'
import { Suspense } from 'react'
import { BackLink } from '@/components/common/BackLink'
import { BackToTop } from '@/components/common/BackToTop'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { ID } from '@/components/common/ID'
import { Text } from '@/components/common/Text'
import { PageSection } from '@/components/layout/PageSection'

import { getAppById, getComponentById, getOrgById } from '@/lib'
import { Builds } from './builds'
import { Config } from './config'
import { Dependencies } from './dependencies'

// NOTE: old layout stuff
import { ErrorBoundary } from 'react-error-boundary'
import {
  BuildComponentButton,
  DashboardContent,
  ErrorFallback,
  Loading,
  Section,
} from '@/components'

export async function generateMetadata({ params }): Promise<Metadata> {
  const {
    ['org-id']: orgId,
    ['app-id']: appId,
    ['component-id']: componentId,
  } = await params
  const [{ data: app }, { data: component }] = await Promise.all([
    getAppById({ appId, orgId }),
    getComponentById({ componentId, orgId }),
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
      getAppById({ appId, orgId }),
      getComponentById({ componentId, orgId }),
      getOrgById({ orgId }),
    ])

  if (error) {
    if (status === 404) {
      notFound()
    } else {
      notFound()
    }
  }

  const containerId = 'app-component-page'
  return org?.features?.['stratus-layout'] ? (
    <PageSection id={containerId} isScrollable className="!p-0 !gap-0">
      {/* old page layout */}
      <div className="p-6 border-b flex justify-between">
        <HeadingGroup>
          <BackLink className="mb-6" />
          <Text variant="base" weight="strong">
            {component?.name}
          </Text>
          <ID>{component.id}</ID>
        </HeadingGroup>

        <div>
          <BuildComponentButton component={component} />
        </div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-12 flex-auto divide-x">
        <div className="divide-y flex flex-col md:col-span-8">
          {component?.dependencies && (
            <Section className="flex-initial" heading="Dependencies">
              <ErrorBoundary fallbackRender={ErrorFallback}>
                <Suspense
                  fallback={
                    <Loading
                      variant="stack"
                      loadingText="Loading component dependencies..."
                    />
                  }
                >
                  <Dependencies component={component} orgId={orgId} />
                </Suspense>
              </ErrorBoundary>
            </Section>
          )}

          <Section heading="Latest config">
            <ErrorBoundary fallbackRender={ErrorFallback}>
              <Suspense
                fallback={
                  <Loading
                    variant="stack"
                    loadingText="Loading component config..."
                  />
                }
              >
                <Config componentId={componentId} orgId={orgId} />
              </Suspense>
            </ErrorBoundary>
          </Section>
        </div>
        <div className="divide-y flex flex-col md:col-span-4">
          <Section heading="Build history">
            <ErrorBoundary fallbackRender={ErrorFallback}>
              <Suspense
                fallback={
                  <Loading variant="stack" loadingText="Loading builds..." />
                }
              >
                <Builds
                  componentId={componentId}
                  orgId={orgId}
                  offset={sp['offset'] || '0'}
                />
              </Suspense>
            </ErrorBoundary>
          </Section>
        </div>
      </div>
      {/* old page layout */}
      <BackToTop containerId={containerId} />
    </PageSection>
  ) : (
    <DashboardContent
      breadcrumb={[
        { href: `/${orgId}/apps`, text: 'Apps' },
        { href: `/${orgId}/apps/${app.id}`, text: app?.name },
        { href: `/${orgId}/apps/${app.id}/components`, text: 'Components' },
        {
          href: `/${orgId}/apps/${app.id}/components/${component.id}`,
          text: component?.name,
        },
      ]}
      heading={component?.name}
      headingUnderline={componentId}
      statues={<BuildComponentButton component={component} />}
    >
      <div className="grid grid-cols-1 md:grid-cols-12 flex-auto divide-x">
        <div className="divide-y flex flex-col md:col-span-8">
          {component?.dependencies && (
            <Section className="flex-initial" heading="Dependencies">
              <ErrorBoundary fallbackRender={ErrorFallback}>
                <Suspense
                  fallback={
                    <Loading
                      variant="stack"
                      loadingText="Loading component dependencies..."
                    />
                  }
                >
                  <Dependencies component={component} orgId={orgId} />
                </Suspense>
              </ErrorBoundary>
            </Section>
          )}

          <Section heading="Latest config">
            <ErrorBoundary fallbackRender={ErrorFallback}>
              <Suspense
                fallback={
                  <Loading
                    variant="stack"
                    loadingText="Loading component config..."
                  />
                }
              >
                <Config componentId={componentId} orgId={orgId} />
              </Suspense>
            </ErrorBoundary>
          </Section>
        </div>
        <div className="divide-y flex flex-col md:col-span-4">
          <Section heading="Build history">
            <ErrorBoundary fallbackRender={ErrorFallback}>
              <Suspense
                fallback={
                  <Loading variant="stack" loadingText="Loading builds..." />
                }
              >
                <Builds
                  componentId={componentId}
                  orgId={orgId}
                  offset={sp['offset'] || '0'}
                />
              </Suspense>
            </ErrorBoundary>
          </Section>
        </div>
      </div>
    </DashboardContent>
  )
}
