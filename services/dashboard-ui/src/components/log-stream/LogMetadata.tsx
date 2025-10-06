'use client'

import { Divider } from '@/components/common/Divider'
import { CodeBlock } from '@/components/common/CodeBlock'
import { Expand } from '@/components/common/Expand'
import { Text } from '@/components/common/Text'
import type { TOTELLog } from '@/types'

export const LogMetadata = ({ log }: { log: TOTELLog }) => {
  return (
    <div className="flex flex-col gap-8 w-full">
      <Divider dividerWord="Metadata" />

      <Expand
        className="self-end border rounded-md"
        id={`${log.id}-json`}
        heading={
          <Text family="mono" variant="subtext">
            View log JSON
          </Text>
        }
      >
        <div className="border-t">
          <CodeBlock language="json">{JSON.stringify(log, null, 2)}</CodeBlock>
        </div>
      </Expand>
    </div>
  )
}
