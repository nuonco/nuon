export default {
  title: 'Workflows/Filters/ShowDriftScan',
}

import { ShowDriftScan } from './ShowDriftScan'

export const Default = () => (
  <div className="p-4">
    <ShowDriftScan showDrifts onChange={() => {}} />
  </div>
)
