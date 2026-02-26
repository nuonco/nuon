import { EmptyGraphic } from './EmptyGraphic'

export const Default = () => <EmptyGraphic />

export const Variants = () => (
  <div className="flex gap-4 items-center">
    <EmptyGraphic variant="404" />
    <EmptyGraphic variant="actions" />
    <EmptyGraphic variant="diagram" />
    <EmptyGraphic variant="history" />
    <EmptyGraphic variant="search" />
    <EmptyGraphic variant="table" />
  </div>
)

export const Small = () => (
  <div className="flex gap-4 items-center">
    <EmptyGraphic variant="404" size="sm" />
    <EmptyGraphic variant="actions" size="sm" />
    <EmptyGraphic variant="diagram" size="sm" />
    <EmptyGraphic variant="history" size="sm" />
    <EmptyGraphic variant="search" size="sm" />
    <EmptyGraphic variant="table" size="sm" />
  </div>
)

export const DarkModeOnly = () => <EmptyGraphic isDarkModeOnly />
