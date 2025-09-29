import type { Metadata } from 'next'
import { Suspense } from 'react'
import { ErrorBoundary } from 'react-error-boundary'
import { DashboardContent, ErrorFallback, Loading, Section } from '@/components'
import { getOrgById } from '@/lib'

import { Installs } from './installs'

export async function generateMetadata({ params }): Promise<Metadata> {
  const { ['org-id']: orgId } = await params
  const { data: org } = await getOrgById({ orgId })

  return {
    title: `Installs | ${org.name} | Nuon`,
  }
}

export default async function InstallsPage({ params, searchParams }) {
  const { ['org-id']: orgId } = await params
  const sp = await searchParams
  return (
    <DashboardContent
      breadcrumb={[{ href: `/${orgId}/installs`, text: 'Installs' }]}
    >
      <Section>
        <ErrorBoundary fallbackRender={ErrorFallback}>
          <Suspense
            fallback={
              <Loading variant="page" loadingText="Loading installs..." />
            }
          >
            <Installs
              orgId={orgId}
              offset={sp['offset'] || '0'}
              q={sp['q'] || ''}
            />
          </Suspense>
        </ErrorBoundary>
      </Section>
    </DashboardContent>
  )
}
