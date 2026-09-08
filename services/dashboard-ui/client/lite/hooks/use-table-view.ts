import {
  useUserPreferences,
  type TCollectionView,
} from '../providers/user-preferences-provider'

export type TTableView = TCollectionView

export const useTableView = () => {
  const { preferences, setPreference } = useUserPreferences()

  return {
    view: preferences.collectionView,
    setView: (view: TTableView) => setPreference('collectionView', view),
  }
}
