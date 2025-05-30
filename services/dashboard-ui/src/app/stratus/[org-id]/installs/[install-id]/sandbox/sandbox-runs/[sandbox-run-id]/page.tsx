import type { FC } from "react"
import { Page, PageHeader, PageNav, Text } from '@/stratus/components'
import type { IPageProps, TSandboxRun } from '@/types'
import { nueQueryData } from '@/utils'

const InstallSandboxRunPage: FC<IPageProps<"org-id" | "install-id" | "sandbox-run-id">> = async ({ params }) => {
  const orgId = params?.['org-id']
  const installId = params?.["install-id"]
  const sandboxRunId = params?.["sandbox-run-id"]
  const { data, error } = await nueQueryData<TSandboxRun>({
    orgId,
    path: `installs/sandbox-runs/${sandboxRunId}`,
  })

  return (
    <div className="flex flex-col px-8 py-6">
      <Text variant="h3" weight="strong">
        {data?.run_type}
      </Text>
      <Text family="mono">
        {data?.id}
      </Text>
    </div>
  )
}

export default InstallSandboxRunPage
