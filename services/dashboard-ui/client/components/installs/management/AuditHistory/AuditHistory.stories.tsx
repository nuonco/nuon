export default {
  title: 'Installs/AuditHistory',
}

import { ModalStory } from '@/components/__stories__/helpers'
import { AuditHistoryModal } from './AuditHistory'

const noop = () => {}

export const Default = () => (
  <ModalStory>
    <AuditHistoryModal error={null} isLoading={false} hasContent={true} onDownload={noop} onDateChange={noop} />
  </ModalStory>
)

export const Loading = () => (
  <ModalStory>
    <AuditHistoryModal error={null} isLoading={true} hasContent={false} onDownload={noop} onDateChange={noop} />
  </ModalStory>
)

export const WithError = () => (
  <ModalStory>
    <AuditHistoryModal error={{ error: 'Unable to load audit logs' }} isLoading={false} hasContent={false} onDownload={noop} onDateChange={noop} />
  </ModalStory>
)
