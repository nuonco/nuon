'use client'

import { Button, type IButtonAsButton } from '@/components/common/Button'
import { CodeBlock } from '@/components/common/CodeBlock'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Banner } from '@/components/common/Banner'
import { ClickToCopyButton } from '@/components/common/ClickToCopy'
import { Skeleton } from '@/components/common/Skeleton'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import { useOrg } from '@/hooks/use-org'
import { useInstall } from '@/hooks/use-install'
import { useSurfaces } from '@/hooks/use-surfaces'
import { useQuery } from '@/hooks/use-query'

interface IViewState {}

export const ViewStateModal = ({ ...props }: IViewState & IModal) => {
  const { removeModal } = useSurfaces()
  const { org } = useOrg()
  const { install } = useInstall()

  const {
    data: state,
    error,
    isLoading,
  } = useQuery<Record<string, any>>({
    path: `/api/orgs/${org?.id}/installs/${install?.id}/state`,
  })

  return (
    <Modal
      heading={
        <Text
          className="inline-flex gap-4 items-center"
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
            <Skeleton height="600px" width="100%" />
          </div>
        ) : state ? (
          <div className="flex flex-col gap-4">
            <div className="flex justify-end">
              <ClickToCopyButton
                textToCopy={JSON.stringify(state, null, 2)}
                className="w-fit"
              />
            </div>
            <CodeBlock 
              language="json"
              className="max-h-[600px]"
            >
              {JSON.stringify(state, null, 2)}
            </CodeBlock>
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