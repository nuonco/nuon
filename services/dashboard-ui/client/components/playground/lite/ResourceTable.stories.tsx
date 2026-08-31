import { ResourceTable, makeRows } from './ResourceTable'

export default {
  title: 'Playground/Lite/ResourceTable',
}

export const Default = () => (
  <div className="p-4">
    <ResourceTable
      headers={['name & id', 'type', 'status', 'updated']}
      rows={makeRows('cmp', 8, 'Component')}
    />
  </div>
)
