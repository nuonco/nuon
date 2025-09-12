import { Pagination, SandboxHistory, Text } from '@/components'
import { getInstallSandboxRuns } from '@/lib'

export const SandboxRuns = async ({
  installId,
  orgId,
  limit = 6,
  offset,
}: {
  installId: string
  orgId: string
  limit?: number
  offset?: string
}) => {
  const { data: sandboxRuns, headers } = await getInstallSandboxRuns({
    orgId,

    installId,
    offset,
    limit,
  })

  const pageData = {
    hasNext: headers?.get('x-nuon-page-next') || 'false',
    offset: headers?.get('x-nuon-page-offset') || '0',
  }

  return sandboxRuns ? (
    <div className="flex flex-col gap-4 w-full">
      <SandboxHistory
        installId={installId}
        orgId={orgId}
        initSandboxRuns={sandboxRuns}
        shouldPoll
      />
      <Pagination
        param="offset"
        pageData={pageData}
        position="center"
        limit={limit}
      />
    </div>
  ) : (
    <Text>Unable to load sandbox history.</Text>
  )
}
