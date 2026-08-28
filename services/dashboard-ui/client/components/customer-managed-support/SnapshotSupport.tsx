import { useState } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useSearchParams } from 'react-router'
import { Badge } from '@/components/common/Badge'
import { Banner } from '@/components/common/Banner'
import { Button } from '@/components/common/Button'
import { Card } from '@/components/common/Card'
import { Hash } from '@/components/common/Hash'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Icon } from '@/components/common/Icon'
import { Link } from '@/components/common/Link'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { Toast } from '@/components/surfaces/Toast'
import { useCustomerManagedSupportSnapshot } from '@/hooks/use-customer-managed-support-snapshot'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { useToast } from '@/hooks/use-toast'
import { uploadCustomerManagedSupportSnapshot } from '@/lib'
import type { TAPIError } from '@/types'
import { formatBytes } from '@/utils/string-utils'

export const CustomerManagedSnapshotSupport = () => {
  const { org } = useOrg()
  const { install } = useInstall()
  const { snapshots, snapshot } = useCustomerManagedSupportSnapshot()
  const { addToast } = useToast()
  const queryClient = useQueryClient()
  const [, setSearchParams] = useSearchParams()
  const [file, setFile] = useState<File>()
  const mutation = useMutation({
    mutationFn: () =>
      uploadCustomerManagedSupportSnapshot({
        orgId: org.id,
        installId: install.id,
        file: file!,
      }),
    onSuccess: (uploaded) => {
      queryClient.invalidateQueries({
        queryKey: ['customer-managed-support-snapshots', org.id, install.id],
      })
      queryClient.setQueryData(
        ['customer-managed-support-snapshot', org.id, install.id, uploaded.id],
        uploaded
      )
      setSearchParams({ snapshot: uploaded.id })
      setFile(undefined)
      addToast(
        <Toast heading="Support snapshot uploaded" theme="success">
          <Text>Captured data for {install.name} is ready to review.</Text>
        </Toast>
      )
    },
  })
  const error = mutation.error as TAPIError | null

  return (
    <>
      <HeadingGroup>
        <Text variant="base" weight="strong">
          Support snapshots
        </Text>
        <Text variant="subtext" theme="neutral">
          Upload and review immutable diagnostic snapshots from this offline
          customer-managed install.
        </Text>
      </HeadingGroup>
      <Card>
        <div className="flex flex-wrap items-start justify-between gap-4">
          <HeadingGroup>
            <Text weight="strong">Upload support snapshot</Text>
            <Text variant="subtext" theme="neutral">
              Use the .tar.zst archive downloaded from the customer portal.
            </Text>
          </HeadingGroup>
          <Button
            variant="primary"
            disabled={!file || mutation.isPending}
            onClick={() => mutation.mutate()}
          >
            <Icon variant="CloudArrowUpIcon" size={16} />
            {mutation.isPending ? 'Uploading snapshot' : 'Upload snapshot'}
          </Button>
        </div>
        <label className="flex cursor-pointer flex-col items-center gap-2 rounded-md border border-dashed p-6 text-center">
          <Icon variant="ArchiveIcon" size={24} />
          <Text weight="strong">Choose a support snapshot</Text>
          <Text variant="subtext" theme="neutral">
            {file
              ? `${file.name} · ${formatBytes(file.size)}`
              : 'Nuon support snapshot (.tar.zst)'}
          </Text>
          <input
            className="sr-only"
            type="file"
            accept=".tar.zst,application/zstd,application/octet-stream"
            onChange={(event) => setFile(event.target.files?.[0])}
          />
        </label>
        {error ? (
          <Banner theme="error">
            <Text weight="strong">Snapshot upload failed</Text>
            <Text variant="subtext">{error.description || error.error}</Text>
          </Banner>
        ) : null}
      </Card>
      <HeadingGroup>
        <Text variant="base" weight="strong">
          Snapshot history
        </Text>
        <Text variant="subtext" theme="neutral">
          Select a capture to view its data across this install.
        </Text>
      </HeadingGroup>
      {snapshots.length ? (
        <div className="flex flex-col overflow-hidden rounded-lg border">
          {snapshots.map((item) => (
            <div
              key={item.id}
              className="flex flex-wrap items-center justify-between gap-4 border-t p-4 first:border-t-0"
            >
              <div className="flex min-w-0 flex-col gap-1">
                <div className="flex flex-wrap items-center gap-2">
                  <Link href={`?snapshot=${item.id}`}>
                    <Time time={item.captured_at} format="long-datetime" />
                  </Link>
                  {item.id === snapshot?.id ? (
                    <Badge theme="info">selected</Badge>
                  ) : null}
                </div>
                <div className="flex flex-wrap items-center gap-3">
                  <Hash hash={item.archive_sha256} />
                  <Text variant="subtext" theme="neutral">
                    {formatBytes(item.archive_size)}
                  </Text>
                  <Text variant="subtext" theme="neutral">
                    {item.manifest.producer.name}
                  </Text>
                </div>
              </div>
              <div className="flex items-center gap-2">
                <Badge theme="success">integrity {item.integrity_status}</Badge>
                <Badge theme="success">
                  association {item.association_status}
                </Badge>
              </div>
            </div>
          ))}
        </div>
      ) : (
        <Banner theme="info">
          No support snapshots uploaded. Upload the first snapshot to review
          captured install data.
        </Banner>
      )}
      {snapshot ? (
        <Card>
          <HeadingGroup>
            <Text weight="strong">Collection report</Text>
            <Text variant="subtext" theme="neutral">
              The customer portal includes only allowlisted data and redacted
              OTEL log fields.
            </Text>
          </HeadingGroup>
          <div className="flex flex-wrap gap-2">
            {snapshot.snapshot.collection.included.map((section) => (
              <Badge key={section}>{section}</Badge>
            ))}
          </div>
          {snapshot.snapshot.collection.unavailable &&
          Object.keys(snapshot.snapshot.collection.unavailable).length ? (
            <Banner theme="warn">
              <Text weight="strong">Some data was unavailable</Text>
              <Text variant="subtext">
                {Object.entries(snapshot.snapshot.collection.unavailable)
                  .map(([name, reason]) => `${name}: ${reason}`)
                  .join(' · ')}
              </Text>
            </Banner>
          ) : null}
        </Card>
      ) : null}
    </>
  )
}
