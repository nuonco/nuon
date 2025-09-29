import { Notice, Text, StacksTable } from '@/components'
import { getInstallStack } from '@/lib'

export const Stacks = async ({
  installId,
  orgId,
}: {
  installId: string
  orgId: string
}) => {
  const { data, error } = await getInstallStack({
    installId,
    orgId,
  })

  return error ? (
    <Notice>Can&apos;t load install stacks: {error?.error}</Notice>
  ) : data?.versions?.length ? (
    <StacksTable stack={data} />
  ) : (
    <Text>No install stacks.</Text>
  )
}
