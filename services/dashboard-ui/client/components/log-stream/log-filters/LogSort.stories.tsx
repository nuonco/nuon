export default {
  title: 'LogStream/LogSort',
}

import { LogSort } from './LogSort'

export const NewestFirst = () => (
  <LogSort
    filters={{
      handleSortToggle: () => {},
      sortStats: { direction: 'desc', isNewestFirst: true, isOldestFirst: false },
    }}
  />
)

export const OldestFirst = () => (
  <LogSort
    filters={{
      handleSortToggle: () => {},
      sortStats: { direction: 'asc', isNewestFirst: false, isOldestFirst: true },
    }}
  />
)
