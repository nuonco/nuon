import { type FC, Suspense } from 'react'
import { ErrorBoundary } from 'react-error-boundary'
import {
  ExampleModal,
  ExampleToast,
  PermToast,
  Header,
  HeadingGroup,
  Link,
  Page,
  Section,
  ScrollableDiv,
  Text,
} from '@/stratus/components'
import type { IPageProps } from '@/types'

const AppsPage: FC<IPageProps<'org-id'>> = async ({ params }) => {
  const { ['org-id']: orgId } = await params

  return (
    <Page
      breadcrumb={{
        baseCrumbs: [
          {
            path: `/stratus/${orgId}`,
            text: 'Home',
          },
          {
            path: `/stratus/${orgId}/apps`,
            text: 'Apps',
          },
        ],
      }}
    >
      <ScrollableDiv>
        <Header>
          <HeadingGroup>
            <Text variant="h3" weight="strong" level={1}>
              Installs
            </Text>
            <Text theme="muted">View your installs here.</Text>
          </HeadingGroup>
        </Header>
        <Section className="divide-y">
          <div className="flex gap-6">
            <ExampleModal />
            <ExampleToast />
            <PermToast />
          </div>

          <ErrorBoundary fallback="An error happened while loading installs">
            <Suspense fallback={'Loading...'}>Apps: TKTKTK</Suspense>
          </ErrorBoundary>
        </Section>
      </ScrollableDiv>
    </Page>
  )
}

export default AppsPage
