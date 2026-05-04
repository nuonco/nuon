export default {
  title: 'VCS Connections/RemoveConnection',
}

import { useState } from 'react'
import { ModalStory } from '@/components/__stories__/helpers'
import { RemoveConnectionModal } from './RemoveConnection'

const noop = () => {}

const Wrapper = ({
  isPending = false,
  error = null,
  initialDeleteGithubApp = false,
}: {
  isPending?: boolean
  error?: any
  initialDeleteGithubApp?: boolean
}) => {
  const [deleteGithubApp, setDeleteGithubApp] = useState(initialDeleteGithubApp)
  return (
    <ModalStory>
      <RemoveConnectionModal
        connectionName="acme-corp"
        isPending={isPending}
        error={error}
        deleteGithubApp={deleteGithubApp}
        onDeleteGithubAppChange={setDeleteGithubApp}
        onSubmit={noop}
      />
    </ModalStory>
  )
}

export const Default = () => <Wrapper />

export const DeleteGithubAppChecked = () => (
  <Wrapper initialDeleteGithubApp={true} />
)

export const Loading = () => <Wrapper isPending={true} />

export const WithError = () => (
  <Wrapper error={{ error: 'Unable to remove VCS connection' }} />
)
