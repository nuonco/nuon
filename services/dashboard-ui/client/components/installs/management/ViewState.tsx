import { useQuery } from '@tanstack/react-query'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { JSONViewer } from "@/components/common/JSONViewer"
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Banner } from '@/components/common/Banner'
import { ClickToCopyButton } from '@/components/common/ClickToCopy'
import { Skeleton } from '@/components/common/Skeleton'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import { useOrg } from '@/hooks/use-org'
import { useInstall } from '@/hooks/use-install'
import { useSurfaces } from '@/hooks/use-surfaces'
import { getInstallState } from '@/lib'

interface IViewState {}

export const ViewStateModal = ({ ...props }: IViewState & IModal) => {
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
    <Modal
      heading={
        <Text
          flex
          className="gap-4"
          variant="h3"
          weight="strong"
        >
          <Icon variant="CodeBlock" size="24" />
          View install state
        </Text>
      }
      className="!max-w-5xl"
      {...props}
    >
      <div className="flex flex-col gap-4">
        {error ? (
          <Banner theme="error">
            {error?.error || 'Unable to load install state.'}
          </Banner>
        ) : null}

        {isLoading ? (
          <div className="flex flex-col gap-4">
            <div className="flex justify-end">
              <Skeleton height="26px" width="26px" />
            </div>
            <Skeleton height="458px" width="100%" />
          </div>
        ) : state ? (
          <div className="flex flex-col gap-4">
            <div className="flex justify-end">
              <ClickToCopyButton
                textToCopy={JSON.stringify(state, null, 2)}
                className="w-fit"
              />
            </div>
            <JSONViewer className="min-h-[458px] max-h-[600px] bg-code" data={state} expanded={1} />
          </div>
        ) : (
          <div className="flex items-center justify-center p-8">
            <Text variant="body" theme="neutral">
              No state data available
            </Text>
          </div>
        )}
      </div>
    </Modal>
  )
}

export const ViewStateButton = ({
  ...props
}: IViewState & IButtonAsButton) => {
  const { addModal } = useSurfaces()
  const modal = <ViewStateModal />

  return (
    <Button
      onClick={() => {
        addModal(modal)
      }}
      {...props}
    >
      View state
      <Icon variant="CodeBlock" />
    </Button>
  )
}
