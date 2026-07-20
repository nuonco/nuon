import { RunbookBuilder } from './RunbookBuilder'

export default { title: 'Runbooks/RunbookBuilder' }

export const Default = () => (
  <RunbookBuilder
    components={[{ id: 'api', name: 'api' }]}
    actions={[{ id: 'backup', name: 'backup' }]}
    runbooks={[]}
  />
)
