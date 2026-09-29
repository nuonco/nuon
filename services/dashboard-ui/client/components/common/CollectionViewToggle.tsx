import { Icon } from './Icon'
import { ToggleButton } from './ToggleButton'

export const COLLECTION_VIEW_STORAGE_KEY = 'nuon:install-resource-collection-view'
export const COLLECTION_VIEW_MODES = ['list', 'grid'] as const
export type TCollectionView = (typeof COLLECTION_VIEW_MODES)[number]

export interface ICollectionViewToggle {
  value: TCollectionView
  onChange: (value: TCollectionView) => void
}

export const CollectionViewToggle = ({
  value,
  onChange,
}: ICollectionViewToggle) => (
  <ToggleButton
    value={value}
    onChange={onChange}
    size="md"
    label="Collection view"
    options={[
      {
        value: 'list',
        label: <Icon variant="ListBulletsIcon" size={16} />,
        ariaLabel: 'List view',
      },
      {
        value: 'grid',
        label: <Icon variant="SquaresFourIcon" size={16} />,
        ariaLabel: 'Grid view',
      },
    ]}
  />
)
