import { ClickToCopyButton } from '@/components/common/ClickToCopy'
import { CodeBlock } from '@/components/common/CodeBlock'
import { Icon } from '@/components/common/Icon'
import { Skeleton } from '@/components/common/Skeleton'
import { Text } from '@/components/common/Text'
import { Banner } from '@/components/common/Banner'
import { Modal, type IModal } from '@/components/surfaces/Modal'

interface IGenerateInstallConfigModal extends IModal {
  content: string | undefined
  error: any
  isLoading: boolean
  onDownload: () => void
}

export const GenerateInstallConfigModal = ({
  content,
  error,
  isLoading,
  onDownload,
  ...props
}: IGenerateInstallConfigModal) => {
  return (
    <Modal
      className="!max-w-5xl"
      heading={
        <Text
          flex
          className="gap-4"
          variant="h3"
          weight="strong"
        >
          <Icon variant="FileCodeIcon" size="24" />
          Generate Install Config
        </Text>
      }
      primaryActionTrigger={
        isLoading || !content
          ? {
              children: (
                <span className="flex items-center gap-2">
                  <Icon variant="Loading" /> Download TOML
                </span>
              ),
              disabled: true,
              variant: 'primary',
            }
          : {
              children: (
                <span className="flex items-center gap-2">
                  <Icon variant="DownloadSimpleIcon" size="18" /> Download TOML
                </span>
              ),
              onClick: onDownload,
              variant: 'primary',
            }
      }
      footerActions={
        !isLoading && content ? (
          <ClickToCopyButton textToCopy={content} className="w-fit" />
        ) : null
      }
      {...props}
    >
      <div className="flex flex-col gap-4">
        {error ? (
          <Banner theme="error">
            {error?.error || 'Unable to load install config TOML'}
          </Banner>
        ) : null}

        {isLoading ? (
          <Skeleton width="100%" height="600px" />
        ) : (
          <CodeBlock language="toml" className="max-h-[600px]">
            {content}
          </CodeBlock>
        )}
      </div>
    </Modal>
  )
}
