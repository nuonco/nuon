'use client'

import { CodeBlock } from '@/components/common/CodeBlock'
import { Duration } from '@/components/common/Duration'
import { JSONViewer } from '@/components/common/JSONViewer'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { InstallActionRunOutputs } from '@/components/actions/InstallActionRunOutputs'
import { InstallActionRunProvider } from '@/providers/install-action-run-provider'
import { toSentenceCase } from '@/utils/string-utils'
import type { IAdhocActionDetails } from './types'

export const AdhocActionDetails = ({ actionRun }: IAdhocActionDetails) => {
  const firstStep = actionRun?.steps?.at(0)
  const adhocConfig = firstStep?.adhoc_config

  return (
    <>
      <div className="flex flex-col gap-2">
        <Text weight="strong">Action step</Text>
        <div className="py-2 px-4 border rounded-md flex flex-col gap-2">
          <div className="flex items-center justify-between">
            <span className="flex items-center gap-2">
              <Status status={firstStep?.status} isWithoutText />
              <Text>{adhocConfig?.name}</Text>
            </span>

            <Text
              className="flex items-center gap-1"
              variant="subtext"
              theme="neutral"
            >
              {toSentenceCase(firstStep?.status)}{' '}
              {firstStep?.execution_duration > 1000000 ? (
                <>
                  in{' '}
                  <Duration
                    variant="subtext"
                    nanoseconds={firstStep?.execution_duration}
                    theme="neutral"
                  />
                </>
              ) : null}
            </Text>
          </div>
        </div>
      </div>

      <div className="flex flex-col gap-2">
        <Text weight="strong">Adhoc action config</Text>
        <div className="flex flex-col gap-4">
          {adhocConfig?.command && (
            <LabeledValue label="Command">
              <CodeBlock language="bash">{adhocConfig.command}</CodeBlock>
            </LabeledValue>
          )}

          {adhocConfig?.inline_contents && (
            <LabeledValue label="Inline contents">
              <CodeBlock language="bash">
                {adhocConfig.inline_contents}
              </CodeBlock>
            </LabeledValue>
          )}

          {(adhocConfig as any)?.timeout && (
            <LabeledValue label="Timeout">
              <Time seconds={(adhocConfig as any).timeout} format="relative" />
            </LabeledValue>
          )}

          {adhocConfig?.env_vars &&
            Object.keys(adhocConfig.env_vars).length > 0 && (
              <LabeledValue label="Environment variables">
                <JSONViewer data={adhocConfig.env_vars} expanded={1} />
              </LabeledValue>
            )}
        </div>
      </div>

      <div className="flex flex-col gap-2">
        <Text weight="strong">Action outputs</Text>
        <InstallActionRunProvider initInstallActionRun={actionRun}>
          <InstallActionRunOutputs />
        </InstallActionRunProvider>
      </div>
    </>
  )
}
