import { useQuery } from '@tanstack/react-query'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import type { IModal } from '@/components/surfaces/Modal'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { useSurfaces } from '@/hooks/use-surfaces'
import { generateCLIInstallConfig } from '@/lib'
import { downloadFileOnClick } from '@/utils/file-download'
import { slugify } from '@/utils/string-utils'
import { GenerateInstallConfigModal } from './GenerateInstallConfig'

export const GenerateInstallConfigModalContainer = ({ ...props }: IModal) => {
  const { org } = useOrg()
  const { install } = useInstall()
  const { removeModal } = useSurfaces()

  const {
    data: config,
    error,
    isLoading,
  } = useQuery({
    queryKey: ['install-generate-cli-config', org.id, install.id],
    queryFn: () =>
      generateCLIInstallConfig({
        installId: install.id,
        orgId: org.id,
      }),
    enabled: !!org?.id && !!install?.id,
  })

  const handleDownload = () => {
    if (config?.content) {
      downloadFileOnClick({
        ...config,
        filename: `${slugify(install.name)}.toml`,
        callback: () => {
          removeModal(props.modalId)
        },
      })
    }
  }

  return (
    <GenerateInstallConfigModal
      content={config?.content}
      error={error}
      isLoading={isLoading}
      onDownload={handleDownload}
      {...props}
    />
  )
}

export const GenerateInstallConfigButton = ({
  ...props
}: IButtonAsButton) => {
  const { addModal } = useSurfaces()
  const modal = <GenerateInstallConfigModalContainer />

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
