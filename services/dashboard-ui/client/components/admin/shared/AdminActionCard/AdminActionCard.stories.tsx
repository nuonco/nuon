export default {
  title: 'Admin/AdminActionCard',
}

import { AdminActionCard } from './AdminActionCard'

export const Default = () => (
  <AdminActionCard
    title="Restart runner"
    description="Restart the runner for this organization."
    variant="default"
    isLoading={false}
    onClick={() => {}}
  />
)

export const Danger = () => (
  <AdminActionCard
    title="Force teardown"
    description="Force teardown of all resources."
    variant="danger"
    isLoading={false}
    onClick={() => {}}
  />
)

export const Loading = () => (
  <AdminActionCard
    title="Restart runner"
    description="Restart the runner for this organization."
    variant="default"
    isLoading={true}
    onClick={() => {}}
  />
)
