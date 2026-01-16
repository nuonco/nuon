'use client'

import { useSearchParams } from 'next/navigation'
import { useEffect } from 'react'
import { Icon } from '@/components/common/Icon'
import { ID } from '@/components/common/ID'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Link } from '@/components/common/Link'
import { Text } from '@/components/common/Text'
import { Modal, IModal } from '@/components/surfaces/Modal'
import { useOrg } from '@/hooks/use-org'
import { useSurfaces } from '@/hooks/use-surfaces'
import type { TVCSConnection } from '@/types'

const VCSAccountLink = ({
  vcs_connection,
}: {
  vcs_connection: TVCSConnection
}) => {
  return (
    <Link
      className="leading-none"
      href={`https://github.com/${vcs_connection?.github_account_name}`}
    >
      <Text variant="subtext" family="mono">
        @{vcs_connection?.github_account_name}
      </Text>
    </Link>
  )
}

const VCSConnectionSuccesModal = (props: IModal) => {
  const searchParams = useSearchParams()
  const { org } = useOrg()
  const vcsId = searchParams?.get('vcs-connected')
  const vcs_connection = org?.vcs_connections?.find((vcs) => vcs?.id === vcsId)

  return (
    <Modal
      heading={
        <div>
          <Text
            className="!flex items-center gap-2"
            variant="h3"
            theme="success"
          >
            <Icon variant="GitHub" />
            GitHub connected to {org?.name}!
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

export const VCSConnectionSuccess = ({}: {}) => {
  const searchParams = useSearchParams()
  const { addModal } = useSurfaces()
  const modal = <VCSConnectionSuccesModal />

  const handleAddModal = () => {
    addModal(modal)
  }

  useEffect(() => {
    if (searchParams?.get('vcs-connected')) {
      handleAddModal()
    }
  }, [])

  return <></>
}
