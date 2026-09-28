export default {
  title: 'Features / Installs / Forms / Resume draft modal',
}

import { ModalStory } from '@/components/__stories__/helpers'
import { ResumeDraftModal } from './ResumeDraftModal'

const recentTimestamp = new Date(Date.now() - 300000).toISOString()
const oldTimestamp = new Date(Date.now() - 86400000 * 2).toISOString()

export const RecentDraft = () => (
  <ModalStory label="Open resume draft modal">
    <ResumeDraftModal
      draftTimestamp={recentTimestamp}
      onResume={() => {}}
      onStartFresh={() => {}}
    />
  </ModalStory>
)

export const OldDraft = () => (
  <ModalStory label="Open old draft modal">
    <ResumeDraftModal
      draftTimestamp={oldTimestamp}
      onResume={() => {}}
      onStartFresh={() => {}}
    />
  </ModalStory>
)
