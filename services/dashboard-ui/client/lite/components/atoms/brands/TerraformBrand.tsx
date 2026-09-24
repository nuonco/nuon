import { SingleColorBrand } from './SingleColorBrand'
import type { IBrandMark } from './types'

const PATH =
  'M1.44 0v7.575l6.561 3.79V3.787zm21.12 4.227l-6.561 3.791v7.574l6.56-3.787zM8.72 4.23v7.575l6.561 3.787V8.018zm0 8.405v7.575L15.28 24v-7.578z'

export const TerraformBrand = (props: IBrandMark) => (
  <SingleColorBrand
    color="light-dark(#7b42bc, #a878e0)"
    path={PATH}
    {...props}
  />
)
