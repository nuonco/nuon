import { PlaceholderGrid } from './PlaceholderGrid'

export default {
  title: 'Playground/Lite/PlaceholderGrid',
}

export const List = () => <PlaceholderGrid rows={8} />

export const TwoColumn = () => <PlaceholderGrid rows={8} columns={2} />

export const Cards = () => (
  <PlaceholderGrid rows={9} columns={3} height="h-[8rem]" />
)
