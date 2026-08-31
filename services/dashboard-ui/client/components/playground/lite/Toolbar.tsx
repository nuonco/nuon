import { Block } from './Block'
import { labelWidth } from './utils'

export interface IToolbar {
  searchWidth?: number
  filters?: string[]
}

export const Toolbar = ({ searchWidth = 280, filters = [] }: IToolbar) => (
  <div className="flex items-center justify-between gap-4">
    <Block
      className="h-[32px]"
      style={{ width: searchWidth }}
      title="Search"
      icon="MagnifyingGlassIcon"
      text="Search"
    />

    {filters.length > 0 && (
      <div className="flex items-center gap-3">
        {filters.map((filter) => (
          <Block
            key={filter}
            className="h-[32px]"
            style={{ width: labelWidth(filter) }}
            title={filter}
            icon="FunnelSimpleIcon"
            text={filter}
          />
        ))}
      </div>
    )}
  </div>
)
