import { Button } from '@/components/common/Button'

export interface IConfigSyncButton {
  isPending?: boolean
  onSync: () => void
}

export const ConfigSyncButton = ({ isPending, onSync }: IConfigSyncButton) => (
  <Button variant="secondary" disabled={isPending} onClick={onSync}>
    {isPending ? 'Syncing config' : 'Sync now'}
  </Button>
)
