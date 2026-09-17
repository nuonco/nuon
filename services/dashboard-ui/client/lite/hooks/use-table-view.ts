import { useEffect, useState } from 'react'
import {
  useUserPreferences,
  type TCollectionView,
} from '../providers/user-preferences-provider'

export type TTableView = TCollectionView

export const useTableView = () => {
  const { preferences } = useUserPreferences()
  const [view, setView] = useState<TTableView>(preferences.collectionView)

  useEffect(() => {
    setView(preferences.collectionView)
  }, [preferences.collectionView])

  return { view, setView }
}
