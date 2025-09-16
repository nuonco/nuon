import type { Metadata } from 'next'
import { redirect } from 'next/navigation'
import { CaretRightIcon } from '@phosphor-icons/react/dist/ssr'
import {
  DashboardContent,
  Link,
  StatusBadge,
  Section,
  Text,
} from '@/components'
import { getOrgById } from '@/lib'

export async function generateMetadata({ params }): Promise<Metadata> {
  const { ['org-id']: orgId } = await params
  const { data: org } = await getOrgById({ orgId })

  return {
    title: `Releases | ${org.name} | Nuon`,
  }
}

export default async function OrgReleases({ params }) {
  const { ['org-id']: orgId } = await params
  const { data: org } = await getOrgById({ orgId })

  if (org?.features?.['org-support']) {
    return (
      <DashboardContent
        breadcrumb={[{ href: `/releases`, text: 'Releases' }]}
        heading={org?.name}
        headingUnderline={org?.id}
        statues={
          <span className="flex flex-col gap-2">
            <Text className="text-cool-grey-600 dark:text-cool-grey-500">
              Status
            </Text>
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
            <Section heading="Changelog" className="flex-initial">
              <Text variant="reg-12">TKTK</Text>
            </Section>
          </div>
          <div className="divide-y flex flex-col flex-auto col-span-4">
            <Section className="flex-initial">
              <div className="flex flex-col gap-3">
                <span>
                  <Text variant="med-18">Introducing Nuon Actions!</Text>
                  <Text
                    className="text-cool-grey-600 dark:text-white/70"
                    variant="reg-12"
                  >
                    Mar 5, 2025
                  </Text>
                </span>
                <Text variant="reg-14" className="!leading-relaxed">
                  Nuon Actions allow you to create automated workflows that can
                  be run in installs. Actions are useful for debugging, running
                  scripts, and implementing health checks.
                </Text>
                <Link
                  href="https://docs.nuon.co/concepts/nuon-actions"
                  target="_blank"
                  className="text-base"
                >
                  Check it out <CaretRightIcon />
                </Link>
              </div>
            </Section>
            <Section className="flex-initial" heading="Recent features">
              <Text variant="reg-12">TKTK</Text>
            </Section>
          </div>
        </div>
      </DashboardContent>
    )
  } else {
    redirect(`/${orgId}/apps`)
  }
}
