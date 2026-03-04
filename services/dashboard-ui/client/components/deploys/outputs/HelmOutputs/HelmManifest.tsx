import { CodeBlock } from '@/components/common/CodeBlock'
import { ClickToCopyButton } from '@/components/common/ClickToCopy'
import { Text } from '@/components/common/Text'

export const HelmManifest = ({ manifest }: { manifest: string }) => {
  return (
    <div className="flex flex-col gap-2">
      <Text weight="strong">Helm manifest</Text>

      <div className="relative">
        <div className="absolute w-fit top-2 right-2">
          <ClickToCopyButton textToCopy={manifest} />
        </div>
        <CodeBlock className="!max-h-fit" language="yml">
          {manifest}
        </CodeBlock>
      </div>
    </div>
  )
}
