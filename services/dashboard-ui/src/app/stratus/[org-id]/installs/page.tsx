import classNames from 'classnames'
import React, { type FC, Suspense } from 'react'
import { ErrorBoundary } from 'react-error-boundary'
import { CaretRight, Cube } from '@phosphor-icons/react/dist/ssr'
import {
  Button,
  Link,
  Header,
  HeaderGroup,
  Page,
  ScrollableDiv,  
  Section,
  Skeleton,
  Status,
  Text,
} from '@/stratus/components'
import type { IPageProps, TInstall } from '@/types'
import { nueQueryData } from '@/utils'

const InstallsPage: FC<IPageProps<'org-id'>> = async ({ params }) => {
  const orgId = params?.['org-id']

  return (
    <Page
      breadcrumb={{
        baseCrumbs: [
          {
            path: `/stratus/${orgId}`,
            text: 'Home',
          },
          {
            path: `/stratus/${orgId}/installs`,
            text: 'Installs',
          },
        ],
      }}
    >
      <ScrollableDiv>
        <Header>
          <HeaderGroup>
            <Text variant="h3" weight="strong" level={1}>
              Installs
            </Text>
            <Text theme="muted">View your installs here.</Text>
          </HeaderGroup>
          <Button className="self-center" variant="primary">
            <Cube />
            Create install
          </Button>
        </Header>
        <Section className="divide-y">
          <ErrorBoundary fallback="An error happened while loading installs">
            <Suspense fallback={<LoadInstallsFallback />}>
              <LoadInstalls orgId={orgId} />
            </Suspense>
          </ErrorBoundary>
        </Section>
      </ScrollableDiv>
    </Page>
  )
}

export default InstallsPage

const InstallsTableHeader: FC = () => (
  <div className="py-4">
    <Grid>
      <Text variant="subtext" theme="muted">
        Name
      </Text>
      <Text variant="subtext" theme="muted">
        Statuses
      </Text>
      <Text variant="subtext" theme="muted">
        App
      </Text>
      <Text variant="subtext" theme="muted">
        Platform
      </Text>
      <span />
    </Grid>
  </div>
)

const LoadInstalls: FC<{ orgId: string }> = async ({ orgId }) => {
  const { data, error } = await nueQueryData<Array<TInstall>>({
    orgId,
    path: `installs`,
  })

  return error ? (
    <Text>Can&apos;t load installs: {error?.error}</Text>
  ) : data?.length ? (
    <>
      <InstallsTableHeader />
      {data?.map((install) => (
        <Grid key={install.id} className="py-4">
          <div className="flex flex-col gap-1">
            <Text variant="base" weight="strong">
              <Link
                className="py-2"
                href={`/stratus/${orgId}/installs/${install?.id}`}
              >
                {install?.name}
              </Link>
            </Text>
            <Text family="mono" variant="subtext" theme="muted">
              {install.id}
            </Text>
          </div>

          <div className="flex flex-col gap-1">
            <Status status={install?.runner_status}>Runner</Status>
            <Status status={install?.sandbox_status}>Sandbox</Status>
            <Status status={install?.composite_component_status}>Components</Status>
          </div>

          <Text>{install?.app?.name}</Text>

          <Text>{install?.app?.cloud_platform}</Text>
          <Link
            className="!p-1 justify-self-end h-fit w-fit"
            href={`/stratus/${orgId}/installs/${install?.id}`}
            variant="ghost"
          >
            <CaretRight />
          </Link>
        </Grid>
      ))}
    </>
  ) : (
    <Text>No installs yet</Text>
  )
}

const Grid: FC<{ children: React.ReactNode; className?: string }> = ({
  children,
  className,
}) => (
  <div
    className={classNames('grid grid-cols-5 gap-4', {
      [`${className}`]: Boolean(className),
    })}
  >
    {children}
  </div>
)

const LoadInstallsFallback: FC = () => (
  <>
    <InstallsTableHeader />
    {[0, 1, 3].map((k) => (
      <Grid key={k} className="py-4">
        <div className="flex flex-col gap-2">
          <Skeleton width="60%" />
          <Skeleton height="12px" />
        </div>

        <div className="flex flex-col gap-2">
          <Skeleton height="12px" width="40%" />
          <Skeleton height="12px" width="40%" />
          <Skeleton height="12px" width="40%" />
        </div>

        <Skeleton width="50%" />

        <Skeleton width="20%" />

        <span />
      </Grid>
    ))}
  </>
)
