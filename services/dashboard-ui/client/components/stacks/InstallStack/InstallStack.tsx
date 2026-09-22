import type { ReactNode } from 'react'
import { Badge } from '@/components/common/Badge'
import { Banner } from '@/components/common/Banner'
import { Card } from '@/components/common/Card'
import { EmptyState } from '@/components/common/EmptyState'
import { Icon } from '@/components/common/Icon'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Skeleton } from '@/components/common/Skeleton'
import { Text } from '@/components/common/Text'
import { CodeBlock } from '@/components/diffs/CodeBlock'
import { SectionHeader } from '@/components/layout/SectionHeader'
import { CustomStackTemplateURL } from '@/components/stacks/CustomStackTemplateURL'
import type { TCustomNestedStack } from '@/types'

export type TInstallNestedStack = {
  id: string
  name: string
  source: 'app config' | 'install override' | 'install only'
  stack: TCustomNestedStack
  type: 'runner' | 'vpc' | 'custom'
}

export interface IInstallStack {
  configAction?: ReactNode
  configError?: string
  configLoading?: boolean
  configVersion?: number
  nestedStacks: TInstallNestedStack[]
  outputs?: string
  outputsAction?: ReactNode
  outputsLoading?: boolean
  stackName?: string
  stackType?: string
  versionsAction?: ReactNode
}

const sourceTheme = {
  'app config': 'neutral',
  'install override': 'info',
  'install only': 'success',
} as const

export const InstallStack = ({
  configAction,
  configError,
  configLoading = false,
  configVersion,
  nestedStacks,
  outputs,
  outputsAction,
  outputsLoading = false,
  stackName,
  stackType,
  versionsAction,
}: IInstallStack) => (
  <div className="flex flex-col gap-6">
    <div className="flex flex-col gap-4">
      <SectionHeader
        title="Current stack"
        actions={
          <>
            {versionsAction}
            {configAction}
          </>
        }
      />

      {configLoading ? (
        <Skeleton height="7rem" width="100%" />
      ) : configError ? (
        <Banner theme="error">{configError}</Banner>
      ) : stackName || stackType || configVersion ? (
        <Card className="!p-4">
          <div className="flex flex-wrap gap-x-8 gap-y-3">
            <LabeledValue label="Name">
              <Text variant="subtext" family="mono">
                {stackName}
              </Text>
            </LabeledValue>
            <LabeledValue label="Type">
              <Badge size="sm">{stackType}</Badge>
            </LabeledValue>
            <LabeledValue label="App config version">
              <Text variant="subtext" family="mono">
                {configVersion}
              </Text>
            </LabeledValue>
          </div>
        </Card>
      ) : (
        <EmptyState
          variant="table"
          emptyTitle="No stack config"
          emptyMessage="This install does not have a stack config."
        />
      )}
    </div>

    {!configLoading && !configError ? (
      <div className="flex flex-col gap-4">
        <SectionHeader
          title="Nested stacks"
          description="Templates included in the install stack."
        />
        {nestedStacks.length ? (
          <div className="flex flex-col gap-2">
            {nestedStacks.map((nestedStack) => (
              <Card key={nestedStack.id} className="!p-4 !gap-3">
                <div className="flex items-center justify-between gap-3 flex-wrap">
                  <div className="flex items-center gap-2 min-w-0">
                    <Icon
                      variant="StackIcon"
                      size={14}
                      className="text-cool-grey-400 shrink-0"
                    />
                    <Text weight="strong" family="mono">
                      {nestedStack.name}
                    </Text>
                    <Badge size="sm">{nestedStack.type}</Badge>
                  </div>
                  <Badge size="sm" theme={sourceTheme[nestedStack.source]}>
                    {nestedStack.source}
                  </Badge>
                </div>
                <CustomStackTemplateURL stack={nestedStack.stack} />
              </Card>
            ))}
          </div>
        ) : (
          <EmptyState
            variant="table"
            emptyTitle="No nested stacks"
            emptyMessage="This stack config does not include nested stack templates."
          />
        )}
      </div>
    ) : null}

    <div className="flex flex-col gap-4">
      <SectionHeader
        title="Outputs"
        description="Values from the last applied stack run."
        actions={outputsAction}
      />
      {outputsLoading && !outputs ? (
        <Skeleton height="12rem" width="100%" />
      ) : outputs ? (
        <CodeBlock
          value={outputs}
          language="json"
          filename="stack-outputs.json"
          copy
          maxHeight={480}
        />
      ) : (
        <EmptyState
          variant="table"
          emptyTitle="No outputs yet"
          emptyMessage="Outputs appear here once the stack has been applied in the cloud account."
        />
      )}
    </div>
  </div>
)
