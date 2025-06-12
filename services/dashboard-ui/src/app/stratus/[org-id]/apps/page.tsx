import { type FC, Suspense } from 'react'
import { ErrorBoundary } from 'react-error-boundary'
import {
  Header,
  HeaderGroup,
  Page,
  Section,
  ScrollableDiv,
  Text,
} from '@/stratus/components'
import type { IPageProps } from '@/types'

const AppsPage: FC<IPageProps<'org-id'>> = ({ params }) => {
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
            path: `/stratus/${orgId}/apps`,
            text: 'Apps',
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
        </Header>
        <Section className="divide-y">
          <ErrorBoundary fallback="An error happened while loading installs">
            <Suspense fallback={'Loading...'}>Apps: TKTKTK</Suspense>
          </ErrorBoundary>
        </Section>
      </ScrollableDiv>
    </Page>
  )
}

export default AppsPage
