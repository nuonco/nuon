import { type FC, Suspense } from 'react'
import { ErrorBoundary } from 'react-error-boundary'
import {
  Markdown,
  ScrollableDiv,
  Section,
  Skeleton,
  Text,
} from '@/stratus/components'
import type { IPageProps, TReadme } from '@/types'
import { nueQueryData } from '@/utils'

const InstallOverviewPage: FC<IPageProps<'org-id' | 'install-id'>> = async ({
  params,
}) => {
  const { ['install-id']: installId, ['org-id']: orgId } = await params

  return (
    <ScrollableDiv>
      <Section>
        <Text variant="base" weight="strong" level={2}>
          Overview
        </Text>

        <ErrorBoundary fallback="Error fetching install README">
          <Suspense fallback={<Skeleton height="800px" />}>
            <LoadInstallReadMe installId={installId} orgId={orgId} />
          </Suspense>
        </ErrorBoundary>
      </Section>
    </ScrollableDiv>
  )
}

export default InstallOverviewPage

const LoadInstallReadMe: FC<{ installId: string; orgId: string }> = async ({
  orgId,
  installId,
}) => {
  const { data, error } = await nueQueryData<TReadme>({
    orgId,
    path: `installs/${installId}/readme`,
  })

  return data ? (
    <Markdown markdownStr={data?.readme as string} />
  ) : (
    <Text>Can&apos;t fetch install README: {error?.error}</Text>
  )
}
