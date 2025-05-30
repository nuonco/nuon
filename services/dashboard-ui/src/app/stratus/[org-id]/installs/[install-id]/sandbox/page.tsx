import { type FC, Suspense } from 'react'
import { ErrorBoundary } from 'react-error-boundary'
import { Link, Text } from '@/stratus/components'
import type { TSandboxRun, IPageProps } from '@/types'
import { nueQueryData } from '@/utils'

const InstallSandboxPage: FC<IPageProps<'org-id' | 'install-id'>> = ({
  params,
}) => {
  const orgId = params?.['org-id']
  const installId = params?.['install-id']

  return (
    <div className="px-8 py-6 flex flex-col">
      <Text variant="base" weight="strong">
        Install sandbox details
      </Text>
      <ErrorBoundary fallback="Error fetching sandbox runs">
        <Suspense fallback="Loaidng sandbox runs...">
          <LoadSandboxRuns installId={installId} orgId={orgId} />
        </Suspense>
      </ErrorBoundary>
    </div>
  )
}

export default InstallSandboxPage

const LoadSandboxRuns: FC<{ installId: string; orgId: string }> = async ({
  installId,
  orgId,
}) => {
  const { data, error } = await nueQueryData<Array<TSandboxRun>>({
    orgId,
    path: `installs/${installId}/sandbox-runs`,
  })

  return (
    <div className="flex flex-col gap-4">
      {error ? <Text>{error?.error}</Text> : null}
      {data?.length ? (
        data?.map((run) => (
          <Link
            key={run?.id}
            href={`/stratus/${orgId}/installs/${installId}/sandbox/sandbox-runs/${run?.id}`}
          >
            {run?.run_type}
          </Link>
        ))
      ) : (
        <Text>No sandbox runs</Text>
      )}
    </div>
  )
}
