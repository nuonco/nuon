export type TSurfaceParam = 'panel' | 'modal'

export interface ISurfaceValue {
  key: string
  resourceId?: string
  value: string
}

export const parseSurfaceValue = (value: string): ISurfaceValue => {
  const separator = value.indexOf(':')
  if (separator < 0) return { key: value, value }

  return {
    key: value.slice(0, separator),
    resourceId: value.slice(separator + 1) || undefined,
    value,
  }
}

export const appendSurfaceValue = (
  search: string,
  param: TSurfaceParam,
  value: string
) => {
  const params = new URLSearchParams(search)
  params.append(param, value)
  return params
}

export const truncateSurfaceValues = (
  search: string,
  param: TSurfaceParam,
  index: number
) => {
  const params = new URLSearchParams(search)
  const retained = params.getAll(param).slice(0, index)
  params.delete(param)
  retained.forEach((value) => params.append(param, value))
  return params
}

export const removeSurfaceValue = (
  search: string,
  param: TSurfaceParam,
  index: number
) => {
  const params = new URLSearchParams(search)
  const retained = params
    .getAll(param)
    .filter((_, valueIndex) => valueIndex !== index)
  params.delete(param)
  retained.forEach((value) => params.append(param, value))
  return params
}
