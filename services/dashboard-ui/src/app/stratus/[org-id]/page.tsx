import type { FC } from 'react'
import { getSession } from '@auth0/nextjs-auth0'
import {
  CaretLeft,
  CaretUp,
  CaretUpDown,
  CaretRight,
  DotsThreeVertical,
  Stack,
} from '@phosphor-icons/react/dist/ssr'
import {
  Button,
  CodeEditor,
  DiffEditor,
  Menu,
  Text,
  Link,
  Tooltip,
  Header,
  HeaderGroup,
  Page,
  Section,
  ScrollableContent,
  Dropdown,
  splitYamlDiff,
} from '@/stratus/components'
import type { IPageProps } from '@/types'
import { nueQueryData } from '@/utils'

const Dashboard: FC<IPageProps<'org-id'>> = async ({ params }) => {
  const orgId = params?.['org-id']
  const { user } = await getSession()
  const { data, error } = await nueQueryData<Record<string, any>>({
    orgId,
    path: `orgs/current`,
  })

  return (
    <Page
      breadcrumb={{
        baseCrumbs: [
          {
            path: `/stratus/${orgId}`,
            text: 'Home',
          },
        ],
      }}
    >
      <Header className="border-b">
        <HeaderGroup>
          <Text variant="h3" weight="strong" level={1}>
            Welcome, {user?.given_name}!
          </Text>
          <Text theme="muted">
            Manage your applications and deployed installs.
          </Text>
        </HeaderGroup>
      </Header>
      <div className="grid md:grid-cols-12 w-full divide-x min-h-full">
        <div className="flex flex-col gap-8 md:col-span-8">
          <ScrollableContent>
            <Section>
              <div className="flex flex-col gap-4">
                <Text variant="h3" weight="strong">
                  Overview
                </Text>

                <div className="grid md:grid-cols-4 rounded-lg border divide-y md:divide-y-0 md:divide-x">
                  <div className="flex flex-col gap-6 p-4">
                    <Text weight="strong" theme="muted">
                      Total installs
                    </Text>

                    <Text variant="h3" weight="strong">
                      10
                    </Text>
                  </div>

                  <div className="flex flex-col gap-6 p-4">
                    <Text weight="strong" theme="muted">
                      Active applications
                    </Text>

                    <Text variant="h3" weight="strong">
                      10
                    </Text>
                  </div>

                  <div className="flex flex-col gap-6 p-4">
                    <Text weight="strong" theme="muted">
                      Active runners
                    </Text>

                    <Text variant="h3" weight="strong">
                      10
                    </Text>
                  </div>

                  <div className="flex flex-col gap-6 p-4">
                    <Text weight="strong" theme="muted">
                      Total installs
                    </Text>

                    <Text variant="h3" weight="strong">
                      10
                    </Text>
                  </div>
                </div>
              </div>
            </Section>
            <Section>
              <Text variant="h3" weight="strong">
                Recent activities
              </Text>
            </Section>
          </ScrollableContent>
        </div>
        <div className="flex flex-col gap-8 md:col-span-4">
          <Section>
            <Text variant="h3" weight="strong">
              Get production ready!
            </Text>
          </Section>
        </div>
      </div>
    </Page>
  )
}

export default Dashboard
