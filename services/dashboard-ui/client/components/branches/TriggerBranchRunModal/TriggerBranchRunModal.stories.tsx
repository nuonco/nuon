export default {
  title: 'Features / Branches / Trigger branch run modal',
}

import { useState } from 'react'
import { ModalStory } from '@/components/__stories__/helpers'
import { Text } from '@/components/common/Text'
import { TriggerBranchRunModal } from './TriggerBranchRunModal'

const noop = () => {}

export const Run = () => {
  const [submittedForce, setSubmittedForce] = useState<boolean>()

  return (
    <div className="flex flex-col gap-4">
      <ModalStory label="Trigger run">
        <TriggerBranchRunModal
          branchName="production"
          planOnly={false}
          isPending={false}
          onConfirm={setSubmittedForce}
        />
      </ModalStory>
      {submittedForce !== undefined ? (
        <Text variant="subtext" theme="neutral">
          Request preview: force = {String(submittedForce)}. No run is launched
          from this story.
        </Text>
      ) : null}
    </div>
  )
}

export const Preview = () => (
  <ModalStory label="Trigger preview">
    <TriggerBranchRunModal
      branchName="production"
      planOnly
      isPending={false}
      onConfirm={noop}
    />
  </ModalStory>
)

export const Pending = () => (
  <ModalStory label="Triggering">
    <TriggerBranchRunModal
      branchName="production"
      planOnly={false}
      isPending
      onConfirm={noop}
    />
  </ModalStory>
)
