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
    <CodeBlock
      value={value}
      language="json"
      filename={filename}
      copy
      maxHeight={640}
      actions={
        onDownload ? (
          <Button
            size="sm"
            variant="icon"
            aria-label="Download"
            tooltipProps={{ tipContent: 'Download' }}
            onClick={onDownload}
          >
            <Icon variant="DownloadSimpleIcon" size={14} />
          </Button>
        ) : null
      }
    />
  )
}
