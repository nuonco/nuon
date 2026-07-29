import { Button } from '@/components/common/Button'
import { EmptyState } from '@/components/common/EmptyState'

interface IInputsNoResults {
  search: string
  hasActiveSearch: boolean
  hasActiveFilters: boolean
  onClearSearch: () => void
  onClearFilters: () => void
  onClearAll: () => void
}

export const InputsNoResults = ({
  search,
  hasActiveSearch,
  hasActiveFilters,
  onClearSearch,
  onClearFilters,
  onClearAll,
}: IInputsNoResults) => (
  <EmptyState
    emptyTitle="No matching inputs"
    emptyMessage={
      hasActiveSearch && hasActiveFilters
        ? `No inputs match "${search}" with the selected filters.`
        : hasActiveSearch
          ? `No inputs match "${search}".`
          : `No inputs match the selected filters.`
    }
    variant="diagram"
    size="sm"
    action={
      <div className="flex items-center gap-2">
        {hasActiveSearch ? (
          <Button size="sm" variant="ghost" onClick={onClearSearch}>
            Clear search
          </Button>
        ) : null}
        {hasActiveFilters ? (
          <Button size="sm" variant="ghost" onClick={onClearFilters}>
            Reset filters
          </Button>
        ) : null}
        {hasActiveSearch && hasActiveFilters ? (
          <Button size="sm" variant="ghost" onClick={onClearAll}>
            Clear all
          </Button>
        ) : null}
      </div>
    }
  />
)
