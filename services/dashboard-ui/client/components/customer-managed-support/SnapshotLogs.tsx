import { useEffect, useState } from 'react'
import { Badge } from '@/components/common/Badge'
import { Button } from '@/components/common/Button'
import { Card } from '@/components/common/Card'
import { EmptyState } from '@/components/common/EmptyState'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { ID } from '@/components/common/ID'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { useCustomerManagedSupportSnapshot } from '@/hooks/use-customer-managed-support-snapshot'
import { CustomerManagedSnapshotContent } from './SnapshotEmpty'
import { CustomerManagedSnapshotLogViewer } from './SnapshotLogViewer'

export const CustomerManagedSnapshotLogs = () => {
  const { snapshot } = useCustomerManagedSupportSnapshot()
  const logs = snapshot?.snapshot.logs ?? []
  const [selectedId, setSelectedId] = useState<string>()
  useEffect(() => {
    if (!logs.some(({ job_id }) => job_id === selectedId))
      setSelectedId(logs[0]?.job_id)
  }, [logs, selectedId])
  const selected = logs.find(({ job_id }) => job_id === selectedId)

  return (
    <CustomerManagedSnapshotContent>
      <HeadingGroup>
        <Text variant="base" weight="strong">
          Job logs
        </Text>
        <Text variant="subtext" theme="neutral">
          View bounded, redacted OTEL job logs included in this support
          snapshot.
        </Text>
      </HeadingGroup>
      {logs.length === 0 ? (
        <EmptyState
          variant="table"
          emptyTitle="No logs captured"
          emptyMessage="Logs will appear after jobs emit OTEL records and the customer captures another snapshot."
        />
      ) : (
        <div className="flex flex-col gap-6">
          <div className="overflow-x-auto rounded-lg border">
            <table className="w-full min-w-[680px] text-sm">
              <thead>
                <tr>
                  {['Job', 'Status', 'Entries', 'Started'].map((label) => (
                    <th
                      key={label}
                      className="bg-cool-grey-100 px-4 py-3 text-left font-normal dark:bg-dark-grey-700"
                    >
                      {label}
                    </th>
                  ))}
                </tr>
              </thead>
              <tbody>
                {logs.map((job) => (
                  <tr
                    key={job.job_id}
                    className={
                      job.job_id === selectedId
                        ? 'bg-cool-grey-50 dark:bg-dark-grey-700'
                        : ''
                    }
                  >
                    <td className="border-t px-4 py-3">
                      <Button
                        variant="ghost"
                        onClick={() => setSelectedId(job.job_id)}
                      >
                        {job.name || job.job_id}
                      </Button>
                    </td>
                    <td className="border-t px-4 py-3">
                      <Status status={job.status || 'unknown'} />
                    </td>
                    <td className="border-t px-4 py-3">
                      {job.entries.length} of {job.total}
                      {job.truncated ? (
                        <Badge className="ml-2" theme="warn">
                          truncated
                        </Badge>
                      ) : null}
                    </td>
                    <td className="border-t px-4 py-3">
                      {job.started_at ? (
                        <Time time={job.started_at} format="relative" />
                      ) : (
                        '—'
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
          {selected ? (
            <Card>
              <div className="flex flex-wrap items-start justify-between gap-4">
                <HeadingGroup>
                  <Text variant="base" weight="strong">
                    {selected.name || 'Job log'}
                  </Text>
                  <ID>{selected.job_id}</ID>
                </HeadingGroup>
                <Status status={selected.status || 'unknown'} variant="badge" />
              </div>
              <CustomerManagedSnapshotLogViewer
                log={selected}
                capturedAt={snapshot?.captured_at}
              />
            </Card>
          ) : null}
        </div>
      )}
    </CustomerManagedSnapshotContent>
  )
}
