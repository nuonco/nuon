import { Icon } from '@/components/common/Icon'
import { ID } from '@/components/common/ID'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import type { TVCSConnection } from '@/types'
import { VCSAccountLink } from '../VCSAccountLink'

interface IVCSConnectionSuccessModal extends IModal {
  orgName?: string
  vcs_connection?: TVCSConnection
}

export const VCSConnectionSuccessModal = ({
  orgName,
  vcs_connection,
  ...props
}: IVCSConnectionSuccessModal) => {
  return (
    <Modal
      heading={
        <div className="flex flex-col">
          <Text
            flex
            className="gap-2"
            variant="h3"
            theme="success"
          >
            <Icon variant="GitHub" />
            GitHub connected to {orgName}!
          </Text>
          <Text variant="subtext">
            <VCSAccountLink vcs_connection={vcs_connection} /> has been
            connected to this organization.
          </Text>
        </div>
      }
      {...props}
    >
      <Text>
        You&apos;re all set! Your repositories can now be used with app configs.
      </Text>

      <div className="flex items-center gap-6">
        <LabeledValue label="Account name">
          <VCSAccountLink vcs_connection={vcs_connection} />
        </LabeledValue>

        <LabeledValue label="Account ID">
          <ID theme="default">{vcs_connection?.github_account_id}</ID>
        </LabeledValue>

        <LabeledValue label="Connection ID">
          <ID theme="default">{vcs_connection?.id}</ID>
        </LabeledValue>
      </div>
    </Modal>
  )
}
