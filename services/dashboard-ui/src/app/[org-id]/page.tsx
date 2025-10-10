import type { Metadata } from 'next'
import { redirect } from 'next/navigation'
import { CaretRightIcon } from '@phosphor-icons/react/dist/ssr'

import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Link } from '@/components/common/Link'
import { PageContent } from '@/components/layout/PageContent'
import { PageGrid } from '@/components/layout/PageGrid'
import { PageHeader } from '@/components/layout/PageHeader'
import { PageLayout } from '@/components/layout/PageLayout'
import { PageSection } from '@/components/layout/PageSection'
import { Text } from '@/components/common/Text'
import { getOrgById } from '@/lib'
import { auth0 } from '@/lib/auth'
import type { TPageProps } from '@/types'

import {
  DashboardContent,
  Link as OldLink,
  StatusBadge,
  Section,
  Text as OldText,
} from '@/components'

export async function generateMetadata({ params }): Promise<Metadata> {
  const { ['org-id']: orgId } = await params
  const { data: org } = await getOrgById({ orgId })

  return {
    title: `${org.name} | Dashboard`,
  }
}

export default async function OrgDashboard({ params }: TPageProps<'org-id'>) {
  const { ['org-id']: orgId } = await params
  const session = await auth0.getSession()
  const { data: org, error } = await getOrgById({ orgId })

  if (error && !org) {
    return (
      <main>
        <h1>Welcome, {session.user.name}!</h1>
        <p>Could not load your organization.</p>
        <div className="flex items-center gap-4">
          <Link href="/">Return home</Link>{' '}
          <Link href="/api/auth/logout">Log out</Link>{' '}
        </div>
      </main>
    )
  }

  if (org?.features?.['org-dashboard']) {
    return org?.features?.['stratus-layout'] ? (
      <PageLayout
        breadcrumb={{
          baseCrumbs: [
            {
              path: `/${orgId}`,
              text: org?.name,
            },
          ],
        }}
        className="divide-y"
        isScrollable
      >
        <PageHeader>
          <HeadingGroup>
            <Text variant="h3" weight="stronger" level={1} role="heading">
              Welcome, {session.user.name}!
            </Text>
            <Text theme="neutral">
              Manage your applications and deployed installs.
            </Text>
          </HeadingGroup>
        </PageHeader>

        <PageContent>
          <PageGrid className="md:divide-x flex-auto">
            <PageSection>
              <Text variant="h3" weight="strong">
                Overview
              </Text>

              <div className="grid md:grid-cols-4 rounded-lg border divide-y md:divide-y-0 md:divide-x">
                <div className="flex flex-col gap-6 p-4">
                  <Text weight="strong" theme="neutral">
                    Total installs
                  </Text>

                  <Text variant="h3" weight="strong">
                    10
                  </Text>
                </div>

                <div className="flex flex-col gap-6 p-4">
                  <Text weight="strong" theme="neutral">
                    Active applications
                  </Text>

                  <Text variant="h3" weight="strong">
                    10
                  </Text>
                </div>

                <div className="flex flex-col gap-6 p-4">
                  <Text weight="strong" theme="neutral">
                    Active runners
                  </Text>

                  <Text variant="h3" weight="strong">
                    10
                  </Text>
                </div>

                <div className="flex flex-col gap-6 p-4">
                  <Text weight="strong" theme="neutral">
                    Total installs
                  </Text>

                  <Text variant="h3" weight="strong">
                    10
                  </Text>
                </div>
              </div>

              <Text variant="h3" weight="strong">
                Recent activity
              </Text>
            </PageSection>
            <PageSection>
              <Text variant="h3" weight="strong">
                Get production ready!
              </Text>
            </PageSection>
          </PageGrid>
        </PageContent>
      </PageLayout>
    ) : (
      <DashboardContent
        breadcrumb={[{ href: `/${orgId}`, text: 'Dashboard' }]}
        heading={org?.name}
        headingUnderline={org?.id}
        statues={
          <span className="flex flex-col gap-2">
            <OldText className="text-cool-grey-600 dark:text-cool-grey-500">
              Status
            </OldText>
            <StatusBadge
              status={org?.status}
              description={org?.status_description}
              descriptionAlignment="right"
            />
          </span>
        }
      >
        <div className="flex-auto md:grid md:grid-cols-12 divide-x">
          <div className="divide-y flex flex-col flex-auto col-span-8">
            <Section heading="Overview" className="flex-initial">
              <OldText variant="reg-12">TKTK</OldText>
            </Section>
            <Section className="flex-initial" heading="Workspaces">
              <OldText variant="reg-12">TKTK</OldText>
            </Section>
          </div>
          <div className="divide-y flex flex-col flex-auto col-span-4">
            <Section className="flex-initial">
              <div className="flex flex-col gap-3">
                <span>
                  <OldText variant="med-18">Introducing Nuon Actions!</OldText>
                  <OldText
                    className="text-cool-grey-600 dark:text-white/70"
                    variant="reg-12"
                  >
                    Mar 5, 2025
                  </OldText>
                </span>
                <OldText variant="reg-14" className="!leading-relaxed">
                  Nuon Actions allow you to create automated workflows that can
                  be run in installs. Actions are useful for debugging, running
                  scripts, and implementing health checks.
                </OldText>
                <OldLink
                  href="https://docs.nuon.co/concepts/nuon-actions"
                  target="_blank"
                  className="text-base"
                >
                  Check it out <CaretRightIcon />
                </OldLink>
              </div>
            </Section>
            <Section className="flex-initial" heading="Recent activity">
              <OldText variant="reg-12">TKTK</OldText>
            </Section>
          </div>
        </div>
      </DashboardContent>
    )
  } else {
    if (org?.features?.['org-runner']) {
      redirect(`/${orgId}/runner`)
    } else {
      redirect(`/${orgId}/apps`)
    }
  }
}
