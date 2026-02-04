import type { Metadata } from 'next'
import { notFound } from 'next/navigation'
import { AsyncBoundary } from '@/components/common/AsyncBoundary'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { getInstall, getOrg } from '@/lib'
import { CurrentInputs } from './inputs'
import { Readme } from './readme'

// NOTE: old install components
import { Loading, Section } from '@/components'

export async function generateMetadata({ params }): Promise<Metadata> {
  const { ['org-id']: orgId, ['install-id']: installId } = await params
  const { data: install } = await getInstall({ installId, orgId })

  return {
    title: `Overview | ${install.name} | Nuon`,
  }
}

export default async function Install({ params }) {
  const { ['org-id']: orgId, ['install-id']: installId } = await params
  const [{ data: install, error, status }, { data: org }] = await Promise.all([
    getInstall({ installId, orgId }),
    getOrg({ orgId }),
  ])

  if (error) {
    if (status === 404) {
      notFound()
    } else {
      notFound()
    }
  }

  return (
    <PageSection className="!pt-0" isScrollable>
      <Breadcrumbs
        breadcrumbs={[
          {
            path: `/${orgId}`,
            text: org?.name,
          },
          {
            path: `/${orgId}/installs`,
            text: 'Installs',
          },
          {
            path: `/${orgId}/installs/${installId}`,
            text: install?.name,
          },
        ]}
      />
      <div className="grid grid-cols-1 md:grid-cols-12 flex-auto divide-x">
        <Section
          heading="README"
          className="md:col-span-8 !p-0"
          headingClassName="px-6 pt-6"
          childrenClassName="overflow-auto px-6 pb-6"
        >
          <AsyncBoundary
            loadingFallback={
              <Loading
                variant="stack"
                loadingText="Loading install README..."
              />
            }
            errorFallback={
              <span className="text-md">Unable to load the REAMDE</span>
            }
          >
            <Readme installId={installId} orgId={orgId} />
          </AsyncBoundary>
        </Section>

        <div className="divide-y flex flex-col col-span-4">
          <Section className="flex-initial">
            <AsyncBoundary
              loadingFallback={
                <Loading
                  variant="stack"
                  loadingText="Loading install inputs..."
                />
              }
              errorFallback={
                <span className="text-md">Unable to load current inputs</span>
              }
            >
              <CurrentInputs installId={installId} orgId={orgId} />
            </AsyncBoundary>
          </Section>
        </div>
      </div>
    </PageSection>
  )
}
