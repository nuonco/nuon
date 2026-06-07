import { Code } from '@/components/common/Code'
import { JSONViewer } from '@/components/common/JSONViewer'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import type { TWorkflow } from '@/types'

type TStatusEntry = NonNullable<
  NonNullable<TWorkflow['status']>['history']
>[number]

interface IWorkflowMetadata {
  workflow: TWorkflow
}

export const WorkflowMetadata = ({ workflow }: IWorkflowMetadata) => {
  const history: TStatusEntry[] = [
    ...(workflow?.status?.history ?? []),
    ...(workflow?.status ? [workflow.status] : []),
  ]

  return (
    <div className="flex flex-col gap-6">
      <div className="flex flex-col gap-2">
        <Text weight="strong">Status history</Text>
        <table className="w-full border-collapse text-left text-sm">
          <thead>
            <tr className="bg-cool-grey-100 dark:bg-dark-grey-700">
              <th className="py-3 px-4 text-left font-normal font-sans rounded-tl-lg">
                Status
              </th>
              <th className="py-3 px-4 text-left font-normal font-sans">Time</th>
              <th className="py-3 px-4 text-left font-normal font-sans rounded-tr-lg">
                Description
              </th>
            </tr>
          </thead>
          <tbody>
            {history.map((status, idx) => (
              <tr key={`${status.created_at_ts}-${idx}`} className="align-top">
                <td className="py-3 px-4 border-t border-cool-grey-200 dark:border-dark-grey-600 whitespace-nowrap">
                  <Status status={status.status} variant="badge" />
                </td>
                <td className="py-3 px-4 border-t border-cool-grey-200 dark:border-dark-grey-600 whitespace-nowrap">
                  <Time
                    seconds={status.created_at_ts}
                    variant="subtext"
                    theme="neutral"
                  />
                </td>
                <td className="py-3 px-4 border-t border-cool-grey-200 dark:border-dark-grey-600">
                  {status.status_human_description ? (
                    <Code className="!text-xs whitespace-pre-wrap break-words">
                      {status.status_human_description}
                    </Code>
                  ) : (
                    <Text variant="subtext" theme="neutral">
                      —
                    </Text>
                  )}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      <div className="flex flex-col gap-2">
        <Text weight="strong">Workflow JSON</Text>
        <div className="border rounded-lg p-4">
          <JSONViewer data={workflow} expanded={1} />
        </div>
      </div>
    </div>
  )
}
