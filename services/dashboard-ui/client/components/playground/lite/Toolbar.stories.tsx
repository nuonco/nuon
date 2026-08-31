import { Toolbar } from './Toolbar'

export default {
  title: 'Playground/Lite/Toolbar',
}

export const Default = () => <Toolbar filters={['Platform']} />

export const MultipleFilters = () => (
  <Toolbar filters={['Platform', 'Status']} />
)

export const SearchOnly = () => <Toolbar />
