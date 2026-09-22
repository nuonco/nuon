import { Banner } from '@/components/common/Banner'
import { Button } from '@/components/common/Button'
import { EmptyState } from '@/components/common/EmptyState'
import { Icon } from '@/components/common/Icon'
import { Skeleton } from '@/components/common/Skeleton'
import { CodeBlock } from '@/components/diffs/CodeBlock'

export interface IInstallState {
  error?: string
  filename?: string
  loading?: boolean
  onDownload?: () => void
  value?: string
}

export const InstallState = ({
  error,
  filename = 'state.json',
  loading = false,
  onDownload,
  value,
}: IInstallState) => {
  if (loading) {
    return <Skeleton height="480px" width="100%" />
  }

  if (error) {
    return <Banner theme="error">{error}</Banner>
  }

  if (!value) {
    return (
      <EmptyState
        emptyTitle="No install state"
        emptyMessage="State appears here after the install has been provisioned."
      />
    )
  }

  return (
    <div className="flex flex-col gap-3">
      {onDownload ? (
        <div className="flex justify-end">
          <Button variant="secondary" onClick={onDownload}>
            <Icon variant="DownloadSimpleIcon" size={16} />
            Download
          </Button>
        </div>
      ) : null}
      <CodeBlock
        value={value}
        language="json"
        filename={filename}
        copy
        maxHeight={640}
      />
    </div>
  )
}
