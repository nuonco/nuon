import { useQuery } from '@tanstack/react-query'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import type { IModal } from '@/components/surfaces/Modal'
import { useOrg } from '@/hooks/use-org'
import { useInstall } from '@/hooks/use-install'
import { useSurfaces } from '@/hooks/use-surfaces'
import { getInstallState } from '@/lib'
import { ViewStateModal } from './ViewState'

export const ViewStateModalContainer = ({ ...props }: IModal) => {
  const { org } = useOrg()
  const { install } = useInstall()

  const {
    data: state,
    error,
    isLoading,
  } = useQuery({
    queryKey: ['install-state', org?.id, install?.id],
    queryFn: () => getInstallState({ orgId: org.id, installId: install.id }),
    enabled: !!org?.id && !!install?.id,
  })

  return (
    <ViewStateModal
      state={state}
      error={error}
      isLoading={isLoading}
      {...props}
    />
  )
}

export const ViewStateButton = ({
  ...props
}: IButtonAsButton) => {
  const { addModal } = useSurfaces()
  const modal = <ViewStateModalContainer />

  return (
    <Button
      onClick={() => {
        addModal(modal)
      }}
      {...props}
    >
      View state
      <Icon variant="CodeBlockIcon" />
    </Button>
  )
}
