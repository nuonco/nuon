export default {
  title: 'Common/SyncedFilter',
}

import { SyncedFilter } from './SyncedFilter'

export const Checked = () => <SyncedFilter showSynced onChange={() => {}} />
export const Unchecked = () => <SyncedFilter showSynced={false} onChange={() => {}} />
