import { SingleColorBrand } from './SingleColorBrand'
import type { IBrandMark } from './types'

const PATH =
  'M0 0v24h24V0zm20.547 20.431H3.448V3.573h17.104V20.43zm-5.155-9.979h3.436v8.255h-3.436zm0-5.16h3.436v3.436h-3.436zm-6.789 9.976V8.732h5.074v-3.44H5.164v13.415h8.513v-3.44Z'

export const OCIBrand = (props: IBrandMark) => (
  <SingleColorBrand
    color="light-dark(#262261, #8b87d1)"
    path={PATH}
    {...props}
  />
)
