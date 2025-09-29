import { ClickToCopyButton, CodeViewer, Text } from '@/components'
import { getInstallComponentOutputs } from '@/lib'

export const LatestOutputs = async ({
  componentId,
  installId,
  orgId,
}: {
  componentId: string
  installId: string
  orgId: string
}) => {
  const { data: outputs, error } = await getInstallComponentOutputs({
    componentId,
    installId,
    orgId,
  })

  return outputs && !error ? (
    <div className="flex flex-col gap-2">
      <div className="flex items-center justify-between">
        <Text variant="med-12">Outputs</Text>
        <ClickToCopyButton textToCopy={JSON.stringify(outputs)} />
      </div>
      <CodeViewer
        initCodeSource={JSON.stringify(outputs, null, 2)}
        language="json"
      />
    </div>
  ) : null
}
