export default {
  title: 'UI / Collection view toggle',
}

import { useState } from 'react'
import {
  CollectionViewToggle,
  type TCollectionView,
} from './CollectionViewToggle'

export const Default = () => {
  const [value, setValue] = useState<TCollectionView>('list')
  return <CollectionViewToggle value={value} onChange={setValue} />
}

export const GridSelected = () => {
  const [value, setValue] = useState<TCollectionView>('grid')
  return <CollectionViewToggle value={value} onChange={setValue} />
}
