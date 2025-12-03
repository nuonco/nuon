'use client'

import { Banner } from '@/components/common/Banner'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { ClickToCopyButton } from '@/components/common/ClickToCopy'
import { CodeBlock } from '@/components/common/CodeBlock'
import { Icon } from '@/components/common/Icon'
import { Skeleton } from '@/components/common/Skeleton'
import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { useQuery } from '@/hooks/use-query'
import { useSurfaces } from '@/hooks/use-surfaces'
import type { TFileResponse } from '@/types'
import { downloadFileOnClick } from '@/utils/file-download'

interface IGenerateInstallConfig {}

export const GenerateInstallConfigModal = ({ ...props }: IGenerateInstallConfig & IModal) => {
  const { org } = useOrg()
  const { install } = useInstall()
  const { removeModal } = useSurfaces()

  const {
    data: config,
    error,
    isLoading,
  } = useQuery<TFileResponse>({
    path: `/api/orgs/${org.id}/installs/${install.id}/generate-cli-config`,
  })

  const handleDownload = () => {
    if (config?.content) {
      downloadFileOnClick({
        ...config,
        callback: () => {
          removeModal(props.modalId)
        },
      })
    }
  }

  return (
    <Modal
      className="!max-w-5xl"
      heading={
        <Text
          className="inline-flex gap-4 items-center"
          variant="h3"
          weight="strong"
        >
          <Icon variant="FileCode" size="24" />
          Generate Install Config
        </Text>
      }
      primaryActionTrigger={
        isLoading || !config?.content
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
                  <Icon variant="DownloadSimple" size="18" /> Download TOML
                </span>
              ),
              onClick: handleDownload,
              variant: 'primary',
            }
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
          <div className="flex flex-col gap-4">
            <div className="flex justify-end">
              <Skeleton width="26px" height="26px" />
            </div>
            <Skeleton width="100%" height="600px" />
          </div>
        ) : (
          <div className="flex flex-col gap-4">
            <div className="flex justify-end">
              <ClickToCopyButton
                textToCopy={config?.content || ''}
                className="w-fit"
              />
            </div>
            <CodeBlock language="json" className="max-h-[600px]">
              {config?.content}
            </CodeBlock>
          </div>
        )}
      </div>
    </Modal>
  )
}

export const GenerateInstallConfigButton = ({
  ...props
}: IGenerateInstallConfig & IButtonAsButton) => {
  const { addModal } = useSurfaces()
  const modal = <GenerateInstallConfigModal />

  return (
    <Button      
      onClick={() => {
        addModal(modal)
      }}
      {...props}
    >
      Generate install config
      <Icon variant="FileCode" />
    </Button>
  )
}
