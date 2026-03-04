'use client'

import { useState } from 'react'
import { Button } from '@/components/common/Button'
import { CreateBranchModal } from './create-branch-modal'

interface ICreateBranchButton {
  appId: string
  orgId: string
}

export const CreateBranchButton = ({ appId, orgId }: ICreateBranchButton) => {
  const [isModalOpen, setIsModalOpen] = useState(false)

  return (
    <>
      <Button onClick={() => setIsModalOpen(true)} variant="primary">
        Create Branch
      </Button>
      <CreateBranchModal
        appId={appId}
        orgId={orgId}
        isOpen={isModalOpen}
        onClose={() => setIsModalOpen(false)}
      />
    </>
  )
}
