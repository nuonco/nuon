export default {
  title: 'Features / Workflows / Filters / Show drift scan',
}

import { ShowDriftScan } from './ShowDriftScan'

export const Default = () => (
  <div className="p-4">
    <ShowDriftScan showDrifts onChange={() => {}} />
  </div>
)
