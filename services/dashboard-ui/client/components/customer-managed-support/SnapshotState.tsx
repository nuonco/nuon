import { useEffect, useState } from 'react'
import { Button } from '@/components/common/Button'
import { EmptyState } from '@/components/common/EmptyState'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Icon } from '@/components/common/Icon'
import { JSONViewer } from '@/components/common/JSONViewer'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { useCustomerManagedSupportSnapshot } from '@/hooks/use-customer-managed-support-snapshot'
import { createFileDownload } from '@/utils/file-download'
import { CustomerManagedSnapshotContent } from './SnapshotEmpty'

export const CustomerManagedSnapshotState = () => {
  const { snapshot } = useCustomerManagedSupportSnapshot()
  const data = snapshot?.snapshot
  const state = data?.state
  const [isCopied, setIsCopied] = useState(false)

  useEffect(() => {
    if (!isCopied) return
    const timeout = setTimeout(() => setIsCopied(false), 2000)
    return () => clearTimeout(timeout)
  }, [isCopied])

  return (
    <CustomerManagedSnapshotContent>
      <div className="flex flex-wrap items-start justify-between gap-4">
        <HeadingGroup>
          <Text variant="base" weight="strong">
            Install state
          </Text>
          <Text variant="subtext" theme="neutral">
            Raw runner status, outputs, and report data included by the
            customer.
          </Text>
        </HeadingGroup>
        {state ? (
          <div className="flex items-center gap-2">
            <Button
              variant="secondary"
              aria-label={isCopied ? 'Copied' : 'Copy state as JSON'}
              title={isCopied ? 'Copied' : 'Copy state as JSON'}
              onClick={() => {
                navigator.clipboard.writeText(JSON.stringify(state, null, 2))
                setIsCopied(true)
              }}
            >
              <Icon variant={isCopied ? 'CheckIcon' : 'CopyIcon'} size="16" />
            </Button>
            <Button
              variant="secondary"
              onClick={() =>
                createFileDownload(
                  JSON.stringify(state, null, 2),
                  'customer-managed-install-state.json',
                  'application/json'
                )
              }
            >
              <Icon variant="DownloadSimpleIcon" size="16" />
              Download
            </Button>
          </div>
        ) : null}
      </div>

      {state ? (
        <>
          <span className="flex items-center gap-2">
            <Text variant="subtext" theme="neutral">
              Captured
            </Text>
            <Time
              time={data?.captured_at}
              format="long-datetime"
              variant="subtext"
            />
          </span>
          <JSONViewer
            className="min-h-[458px] max-h-[80vh] bg-code"
            data={state}
            expanded={1}
          />
        </>
      ) : data?.include_state === false ? (
        <EmptyState
          variant="table"
          emptyTitle="State not included"
          emptyMessage="The customer chose not to include raw state in this support snapshot."
        />
      ) : data?.include_state ? (
        <EmptyState
          variant="table"
          emptyTitle="State unavailable"
          emptyMessage="The customer selected raw state, but it was unavailable when this snapshot was created."
        />
      ) : (
        <EmptyState
          variant="table"
          emptyTitle="State not captured"
          emptyMessage="This support snapshot was created by a runner that did not record the state selection."
        />
      )}
    </CustomerManagedSnapshotContent>
  )
}
