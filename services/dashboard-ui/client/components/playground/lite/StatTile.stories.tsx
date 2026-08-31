import { StatTile } from './StatTile'

export default {
  title: 'Playground/Lite/StatTile',
}

export const Default = () => <StatTile label="Installs" />

export const Row = () => (
  <div className="grid grid-cols-4 gap-4">
    <StatTile label="Apps" valueWidth={40} />
    <StatTile label="Installs" valueWidth={56} />
    <StatTile label="Deploys today" valueWidth={72} />
    <StatTile label="Failing" valueWidth={32} />
  </div>
)
