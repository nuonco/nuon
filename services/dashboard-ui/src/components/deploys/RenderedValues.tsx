import { KeyValueList } from '@/components/common/KeyValueList'
import type { TKeyValue } from '@/types'
import { objectToKeyValueArray } from '@/utils/data-utils'

type TRenderedValue = Record<string, string>

export const RenderedValues = ({
  values,
}: {
  values: TRenderedValue | TRenderedValue[]
}) => {
  let vals: TKeyValue[]
  if (Array.isArray(values)) {
    vals = values?.map((v) => ({ key: v?.name, value: v?.value }))
  } else {
    vals = objectToKeyValueArray(values)
  }

  return <KeyValueList values={vals} />
}
