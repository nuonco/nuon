import type { TComponent } from '@/types'
import { buildComponent } from './build-component'

export const buildComponents = ({
  components,
  orgId,
}: {
  components: TComponent[]
  orgId: string
}) => {
  return Promise.all(
    components.map(({ id, name }) =>
      buildComponent({
        componentId: id,
        orgId,
      }).then((res) =>
        res?.error
          ? { ...res, error: { ...res.error, meta: { name, id } } }
          : res
      )
    )
  )
}
