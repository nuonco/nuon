export default {
  title: 'Spotlight/SpotlightResultItem',
}

import { SpotlightResultItem } from './SpotlightResultItem'

const mockResult = {
  path: '/org-1/installs/install-1',
  label: 'production',
  subtitle: 'Install',
  icon: 'Cube' as const,
  tag: 'install' as any,
}

export const Default = () => (
  <div className="w-80 p-2 border rounded">
    <SpotlightResultItem
      result={mockResult}
      isActive={false}
      index={0}
      onSelect={() => {}}
      onHover={() => {}}
    />
  </div>
)

export const Active = () => (
  <div className="w-80 p-2 border rounded">
    <SpotlightResultItem
      result={mockResult}
      isActive={true}
      index={0}
      onSelect={() => {}}
      onHover={() => {}}
    />
  </div>
)

export const Command = () => (
  <div className="w-80 p-2 border rounded">
    <SpotlightResultItem
      result={{ ...mockResult, label: 'Create install', subtitle: undefined, tag: 'command' }}
      isActive={false}
      index={1}
      onSelect={() => {}}
      onHover={() => {}}
    />
  </div>
)

export const WithoutSubtitle = () => (
  <div className="w-80 p-2 border rounded">
    <SpotlightResultItem
      result={{ ...mockResult, subtitle: undefined, tag: undefined }}
      isActive={false}
      index={0}
      onSelect={() => {}}
      onHover={() => {}}
    />
  </div>
)
