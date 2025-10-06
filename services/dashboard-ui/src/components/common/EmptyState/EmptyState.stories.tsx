import { EmptyState } from './EmptyState'

export const Default = () => <EmptyState />

export const CustomText = () => (
  <EmptyState
    emptyTitle="No data available"
    emptyMessage="Try adjusting your filters or check back later"
  />
)

export const Variants = () => (
  <div className="flex flex-wrap gap-8">
    <EmptyState
      variant="404"
      emptyTitle="Page not found"
      emptyMessage="The page you're looking for doesn't exist"
    />
    <EmptyState
      variant="actions"
      emptyTitle="No actions"
      emptyMessage="No actions available for this resource"
    />
    <EmptyState
      variant="diagram"
      emptyTitle="No diagram"
      emptyMessage="No diagram data available"
    />
    <EmptyState
      variant="history"
      emptyTitle="No history"
      emptyMessage="No history records found"
    />
    <EmptyState
      variant="search"
      emptyTitle="No results"
      emptyMessage="Try adjusting your search terms"
    />
    <EmptyState
      variant="table"
      emptyTitle="No data"
      emptyMessage="No table data available"
    />
  </div>
)

export const Small = () => (
  <div className="flex flex-wrap gap-8">
    <EmptyState
      variant="404"
      size="sm"
      emptyTitle="Not found"
      emptyMessage="Page not found"
    />
    <EmptyState
      variant="search"
      size="sm"
      emptyTitle="No results"
      emptyMessage="Try different keywords"
    />
    <EmptyState
      variant="table"
      size="sm"
      emptyTitle="Empty"
      emptyMessage="No data"
    />
  </div>
)

export const DarkModeOnly = () => (
  <EmptyState
    isDarkModeOnly
    emptyTitle="Dark mode component"
    emptyMessage="This component is optimized for dark mode"
  />
)
