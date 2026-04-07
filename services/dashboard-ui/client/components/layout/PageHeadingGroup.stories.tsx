export default {
  title: 'Layout/PageHeadingGroup',
}

import { PageHeadingGroup } from './PageHeadingGroup'

export const Default = () => (
  <PageHeadingGroup title="Page title" />
)

export const WithSubtitle = () => (
  <PageHeadingGroup
    title="Installs"
    subtitle="Manage all your installs across different apps"
  />
)

export const WithBackLink = () => (
  <PageHeadingGroup
    title="Install details"
    subtitle="View and manage this install"
    showBackLink
  />
)
