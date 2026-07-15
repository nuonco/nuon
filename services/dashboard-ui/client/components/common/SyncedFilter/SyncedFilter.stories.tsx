export default {
  title: 'Common/SyncedFilter',
}

import { SyncedFilter } from './SyncedFilter'

export const Checked = () => <SyncedFilter syncedOnly onChange={() => {}} />
export const Unchecked = () => <SyncedFilter syncedOnly={false} onChange={() => {}} />
