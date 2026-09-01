export default {
  title: 'Branches/EditBranchNameModal',
}

import { ModalStory } from '@/components/__stories__/helpers'
import { EditBranchNameModal } from './EditBranchNameOnlyModal'

export const EditName = () => (
  <ModalStory>
    <EditBranchNameModal
      branchName="production"
      isPending={false}
      error={null}
      onSubmit={() => {}}
      onCancel={() => {}}
    />
  </ModalStory>
)
