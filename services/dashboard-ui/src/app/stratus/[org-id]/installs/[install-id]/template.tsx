'use client'

import { usePathname } from 'next/navigation'
import {
  CalendarDots,
  HouseSimple,
  ShippingContainer,
  SneakerMove,
  Cards,
  Stack,
  TerminalWindow,
} from '@phosphor-icons/react/dist/ssr'
import {
  Header,
  HeaderGroup,
  PageLayout,
  Page,
  PageNav,
  Text,
} from '@/stratus/components'
import { useInstall, useOrg } from '@/stratus/context'

export default function Template({ children }: { children: React.ReactNode }) {
  const pathName = usePathname()
  const { org } = useOrg()
  const { install } = useInstall()
  const isThirdLevel = pathName.split('/').length > 6

  return isThirdLevel ? (
    <Page
      breadcrumb={{
        baseCrumbs: [
          {
            path: `/stratus/${org?.id}`,
            text: 'Home',
          },
          {
            path: `/stratus/${org?.id}/installs`,
            text: 'Installs',
          },
          {
            path: `/stratus/${org?.id}/installs/${install?.id}`,
            text: install?.name || 'Install',
          },
        ],
      }}
    >
      {children}
    </Page>
  ) : (
    <Page
      breadcrumb={{
        baseCrumbs: [
          {
            path: `/stratus/${org?.id}`,
            text: 'Home',
          },
          {
            path: `/stratus/${org?.id}/installs`,
            text: 'Installs',
          },
          {
            path: `/stratus/${org?.id}/installs/${install?.id}`,
            text: install?.name || 'Install',
          },
        ],
      }}
    >
      <Header className="border-b">
        <HeaderGroup>
          <Text variant="h3" weight="strong" level={1}>
            {install?.name}
          </Text>
          <Text family="mono" variant="subtext" theme="muted">
            {install?.id}
          </Text>
        </HeaderGroup>
        <div>Install statuses</div>
      </Header>
      <PageLayout>
        <PageNav
          basePath={`/stratus/${org?.id}/installs/${install?.id}`}
          links={[
            {
              path: `/`,
              icon: <HouseSimple />,
              text: 'Overview',
            },
            {
              path: `/runner`,
              icon: <SneakerMove />,
              text: 'Runner',
            },
            {
              path: `/sandbox`,
              icon: <ShippingContainer />,
              text: 'Sandbox',
            },
            {
              path: `/stacks`,
              icon: <Stack />,
              text: 'Stacks',
            },
            {
              path: `/components`,
              icon: <Cards />,
              text: 'Components',
            },
            {
              path: `/actions`,
              icon: <TerminalWindow />,
              text: 'Actions',
            },
            {
              path: `/history`,
              icon: <CalendarDots />,
              text: 'History',
            },
          ]}
        />
        {children}
      </PageLayout>
    </Page>
  )
}
