import { ClickToCopyButton } from '@/components/common/ClickToCopy'
import { KeyValueList } from '@/components/common/KeyValueList'
import { Text } from '@/components/common/Text'
import type { TInstallStackVersionRun } from '@/types'
import { objectToKeyValueArray } from '@/utils/data-utils'

export const StackOutputs = ({ runs }: { runs: TInstallStackVersionRun[] }) => {
  return runs?.map((run, i) => (
    <div key={run?.id} className="flex flex-col gap-4">
      <Text className="!flex items-center justify-between">
        <Text variant="body" weight="strong">
          Run {i + 1}
        </Text>
        <ClickToCopyButton
          className="w-fit self-end"
          textToCopy={JSON.stringify(run.data_contents)}
        />
      </Text>
      <div className="overflow-auto max-h-[600px]">
        <KeyValueList values={objectToKeyValueArray(run?.data)} />
      </div>
    </div>
  ))
}
