import type { Metadata } from 'next'
import { Suspense, type FC } from 'react'
import { ErrorBoundary } from 'react-error-boundary'
import { withPageAuthRequired } from '@auth0/nextjs-auth0'
import {
  BuildComponentButton,
  ComponentBuildHistory,
  ComponentConfiguration,
  ComponentDependencies,
  DashboardContent,
  ErrorFallback,
  Loading,
  Section,
  Text,
} from '@/components'
import {
  getApp,
  getComponent,
  getComponentBuilds,
  getComponentConfig,
} from '@/lib'
import type { TComponent } from '@/types'
import { nueQueryData } from '@/utils'

export async function generateMetadata({ params }): Promise<Metadata> {
  const appId = params?.['app-id'] as string
  const componentId = params?.['component-id'] as string
  const orgId = params?.['org-id'] as string
  const [app, component] = await Promise.all([
    getApp({ appId, orgId }),
    getComponent({ componentId, orgId }),
  ])

  return {
    title: `${app.name} | ${component.name}`,
  }
}

export default withPageAuthRequired(async function AppComponent({ params }) {
  const appId = params?.['app-id'] as string
  const componentId = params?.['component-id'] as string
  const orgId = params?.['org-id'] as string

  const [app, component] = await Promise.all([
    getApp({ appId, orgId }),
    getComponent({ componentId, orgId }),
  ])

  return (
    <DashboardContent
      breadcrumb={[
        { href: `/${orgId}/apps`, text: 'Apps' },
        { href: `/${orgId}/apps/${app.id}`, text: app.name },
        { href: `/${orgId}/apps/${app.id}/components`, text: 'Components' },
        {
          href: `/${orgId}/apps/${app.id}/components/${component.id}`,
          text: component.name,
        },
      ]}
      heading={component.name}
      headingUnderline={component.id}
      statues={<BuildComponentButton componentName={component?.name} />}
    >
      <div className="grid grid-cols-1 md:grid-cols-12 flex-auto divide-x">
        <div className="divide-y flex flex-col md:col-span-8">
          {component.dependencies && (
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
                  <LoadComponentDependencies
                    component={component}
                    orgId={orgId}
                  />
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
                <LoadComponentConfig componentId={componentId} orgId={orgId} />
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
                <LoadComponentBuilds
                  appId={appId}
                  componentId={componentId}
                  orgId={orgId}
                />
              </Suspense>
            </ErrorBoundary>
          </Section>
        </div>
      </div>
    </DashboardContent>
  )
})

const LoadComponentBuilds: FC<{
  appId: string
  componentId: string
  orgId: string
}> = async ({ appId, componentId, orgId }) => {
  const builds = await getComponentBuilds({ componentId, orgId })
  return (
    <ComponentBuildHistory
      appId={appId}
      componentId={componentId}
      initBuilds={builds}
      orgId={orgId}
      shouldPoll
    />
  )
}

const LoadComponentConfig: FC<{ componentId: string; orgId: string }> = async ({
  componentId,
  orgId,
}) => {
  const componentConfig = await getComponentConfig({ componentId, orgId })
  return <ComponentConfiguration config={componentConfig} isNotTruncated />
}

const LoadComponentDependencies: FC<{
  component: TComponent
  orgId: string
}> = async ({ component, orgId }) => {
  const { data, error } = await nueQueryData<Array<TComponent>>({
    orgId,
    path: `components/${component?.id}/dependencies`,
  })

  return (
    <div className="flex items-center gap-4">
      {error ? (
        <Text>{error?.error}</Text>
      ) : (
        <ComponentDependencies deps={data} name={component?.name} />
      )}
    </div>
  )
}
