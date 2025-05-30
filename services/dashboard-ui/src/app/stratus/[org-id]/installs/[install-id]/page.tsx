import { type FC, Suspense } from 'react'
import { ErrorBoundary } from 'react-error-boundary'
import { Markdown, Text } from '@/stratus/components'
import type { IPageProps, TReadme } from '@/types'
import { nueQueryData } from '@/utils'

const InstallOverviewPage: FC<IPageProps<'org-id' | 'install-id'>> = async ({
  params,
}) => {
  const orgId = params?.['org-id']
  const installId = params?.['install-id']

  return (
    <div className="px-8 py-6 flex-auto w-full flex flex-col overflow-scroll">
      <Text variant="base" weight="strong">
        Install overview
      </Text>

      <ErrorBoundary fallback="Error fetching install README">
        <Suspense fallback="Loading...">
          <LoadInstallReadMe installId={installId} orgId={orgId} />
        </Suspense>
      </ErrorBoundary>
    </div>
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
    <Text>Can&apso;t fetch install README: {error?.error}</Text>
  )
}
