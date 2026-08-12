export default {
  title: 'Branches/Filters/ShowPreviewRuns',
}

import { ShowPreviewRuns } from './ShowPreviewRuns'

export const Checked = () => (
  <div className="p-4">
    <ShowPreviewRuns showPreviews onChange={() => {}} />
  </div>
)

export const Unchecked = () => (
  <div className="p-4">
    <ShowPreviewRuns showPreviews={false} onChange={() => {}} />
  </div>
)
