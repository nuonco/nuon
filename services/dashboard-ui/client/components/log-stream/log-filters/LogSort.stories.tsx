export default {
  title: 'LogStream/LogSort',
}

import { LogSort } from './LogSort'

export const NewestFirst = () => (
  <LogSort
    filters={{
      handleSortToggle: () => {},
      sortStats: { isNewestFirst: true },
    }}
  />
)

export const OldestFirst = () => (
  <LogSort
    filters={{
      handleSortToggle: () => {},
      sortStats: { isNewestFirst: false },
    }}
  />
)
