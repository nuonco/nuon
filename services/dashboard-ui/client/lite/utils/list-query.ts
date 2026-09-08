export type TListQueryKeyValue =
  | string
  | number
  | boolean
  | null
  | readonly TListQueryKeyValue[]

export interface IListQueryCodec<T> {
  parse: (values: string[]) => T
  serialize: (value: T) => string[]
  toQueryKey: (value: T) => TListQueryKeyValue
}

export interface IListQueryParameter<T> {
  param: string
  codec: IListQueryCodec<T>
  resetsOffset?: boolean
}

export type TListQueryParameters = Record<string, IListQueryParameter<any>>

export type TListQueryValues<T extends TListQueryParameters> = {
  [K in keyof T]: T[K] extends IListQueryParameter<infer V> ? V : never
}

export const stringQueryParameter = (
  param: string,
  {
    defaultValue = '',
    resetsOffset = true,
  }: {
    defaultValue?: string
    resetsOffset?: boolean
  } = {}
): IListQueryParameter<string> => ({
  param,
  resetsOffset,
  codec: {
    parse: (values) => values.at(-1) ?? defaultValue,
    serialize: (value) =>
      value === defaultValue || value === '' ? [] : [value],
    toQueryKey: (value) => value,
  },
})

export const booleanQueryParameter = (
  param: string,
  {
    defaultValue = false,
    resetsOffset = true,
  }: {
    defaultValue?: boolean
    resetsOffset?: boolean
  } = {}
): IListQueryParameter<boolean> => ({
  param,
  resetsOffset,
  codec: {
    parse: (values) => {
      const value = values.at(-1)
      if (value === 'true') return true
      if (value === 'false') return false
      return defaultValue
    },
    serialize: (value) =>
      value === defaultValue ? [] : [value ? 'true' : 'false'],
    toQueryKey: (value) => value,
  },
})

export const enumQueryParameter = <T extends string>(
  param: string,
  options: readonly T[],
  {
    defaultValue,
    resetsOffset = true,
  }: {
    defaultValue: T
    resetsOffset?: boolean
  }
): IListQueryParameter<T> => ({
  param,
  resetsOffset,
  codec: {
    parse: (values) => {
      const value = values.at(-1)
      return value && options.includes(value as T) ? (value as T) : defaultValue
    },
    serialize: (value) => (value === defaultValue ? [] : [value]),
    toQueryKey: (value) => value,
  },
})

const parseSetValues = (values: string[], separator?: string) =>
  new Set(
    values
      .flatMap((value) => (separator ? value.split(separator) : [value]))
      .map((value) => value.trim())
      .filter(Boolean)
  )

const setQueryKey = (value: Set<string>) => [...value].sort()

export const commaSetQueryParameter = <T extends string = string>(
  param: string,
  {
    defaultValue = [],
    resetsOffset = true,
  }: {
    defaultValue?: readonly T[]
    resetsOffset?: boolean
  } = {}
): IListQueryParameter<Set<T>> => {
  const defaultKey = [...defaultValue].sort()

  return {
    param,
    resetsOffset,
    codec: {
      parse: (values) =>
        values.length
          ? (parseSetValues(values, ',') as Set<T>)
          : new Set(defaultValue),
      serialize: (value) => {
        const sorted = [...value].sort()
        if (JSON.stringify(sorted) === JSON.stringify(defaultKey)) return []
        return sorted.length ? [sorted.join(',')] : []
      },
      toQueryKey: (value) => setQueryKey(value),
    },
  }
}

export const repeatedSetQueryParameter = <T extends string = string>(
  param: string,
  {
    defaultValue = [],
    resetsOffset = true,
  }: {
    defaultValue?: readonly T[]
    resetsOffset?: boolean
  } = {}
): IListQueryParameter<Set<T>> => {
  const defaultKey = [...defaultValue].sort()

  return {
    param,
    resetsOffset,
    codec: {
      parse: (values) =>
        values.length
          ? (parseSetValues(values) as Set<T>)
          : new Set(defaultValue),
      serialize: (value) => {
        const sorted = [...value].sort()
        return JSON.stringify(sorted) === JSON.stringify(defaultKey)
          ? []
          : sorted
      },
      toQueryKey: (value) => setQueryKey(value),
    },
  }
}

export const offsetQueryParameter = (
  param = 'offset'
): IListQueryParameter<number> => ({
  param,
  resetsOffset: false,
  codec: {
    parse: (values) => {
      const value = Number(values.at(-1))
      return Number.isFinite(value) && value >= 0 ? Math.floor(value) : 0
    },
    serialize: (value) => (value > 0 ? [String(Math.floor(value))] : []),
    toQueryKey: (value) => value,
  },
})

export const readQueryParameter = <T>(
  searchParams: URLSearchParams,
  parameter: IListQueryParameter<T>
) => parameter.codec.parse(searchParams.getAll(parameter.param))

export const writeQueryParameter = <T>(
  searchParams: URLSearchParams,
  parameter: IListQueryParameter<T>,
  value: T
) => {
  searchParams.delete(parameter.param)
  parameter.codec
    .serialize(value)
    .forEach((serialized) => searchParams.append(parameter.param, serialized))
}
