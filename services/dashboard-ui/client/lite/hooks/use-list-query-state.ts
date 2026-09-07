import { useCallback, useMemo } from 'react'
import { useSearchParams } from 'react-router'
import {
  offsetQueryParameter,
  readQueryParameter,
  stringQueryParameter,
  writeQueryParameter,
  type IListQueryParameter,
  type TListQueryParameters,
  type TListQueryValues,
} from '../lib/list-query'

const DEFAULT_SEARCH = stringQueryParameter('q')
const DEFAULT_OFFSET = offsetQueryParameter()

export interface IListQueryStateConfig<
  TFilters extends TListQueryParameters = Record<string, never>,
> {
  pageSize: number
  search?: IListQueryParameter<string> | false
  offset?: IListQueryParameter<number>
  filters?: TFilters
}

export interface IListQueryState<TFilters extends TListQueryParameters> {
  search: string
  offset: number
  pageSize: number
  filters: TListQueryValues<TFilters>
  queryKey: readonly unknown[]
  setSearch: (value: string) => void
  setOffset: (value: number) => void
  setFilter: <K extends keyof TFilters>(
    key: K,
    value: TListQueryValues<TFilters>[K]
  ) => void
  setFilters: (values: Partial<TListQueryValues<TFilters>>) => void
  resetFilters: () => void
}

const equalQueryValues = <T>(
  parameter: IListQueryParameter<T>,
  left: T,
  right: T
) =>
  JSON.stringify(parameter.codec.toQueryKey(left)) ===
  JSON.stringify(parameter.codec.toQueryKey(right))

export const useListQueryState = <
  TFilters extends TListQueryParameters = Record<string, never>,
>({
  pageSize,
  search: searchConfig = DEFAULT_SEARCH,
  offset: offsetConfig = DEFAULT_OFFSET,
  filters: filterConfig = {} as TFilters,
}: IListQueryStateConfig<TFilters>): IListQueryState<TFilters> => {
  const [searchParams, setSearchParams] = useSearchParams()

  const search =
    searchConfig === false ? '' : readQueryParameter(searchParams, searchConfig)
  const offset = readQueryParameter(searchParams, offsetConfig)

  const filters = useMemo(
    () =>
      Object.fromEntries(
        Object.entries(filterConfig).map(([key, parameter]) => [
          key,
          readQueryParameter(searchParams, parameter),
        ])
      ) as TListQueryValues<TFilters>,
    [filterConfig, searchParams]
  )

  const queryKey = useMemo(
    () => [
      search,
      offset,
      pageSize,
      ...Object.keys(filterConfig)
        .sort()
        .flatMap((key) => [
          key,
          filterConfig[key].codec.toQueryKey(filters[key]),
        ]),
    ],
    [filterConfig, filters, offset, pageSize, search]
  )

  const updateParams = useCallback(
    (
      update: (next: URLSearchParams, current: URLSearchParams) => void,
      replace: boolean
    ) => {
      setSearchParams(
        (current) => {
          const next = new URLSearchParams(current)
          update(next, current)
          return next
        },
        { replace }
      )
    },
    [setSearchParams]
  )

  const setSearch = useCallback(
    (value: string) => {
      if (searchConfig === false) return
      updateParams((next) => {
        writeQueryParameter(next, searchConfig, value)
        writeQueryParameter(next, offsetConfig, 0)
      }, true)
    },
    [offsetConfig, searchConfig, updateParams]
  )

  const setOffset = useCallback(
    (value: number) => {
      updateParams(
        (next) => writeQueryParameter(next, offsetConfig, Math.max(0, value)),
        false
      )
    },
    [offsetConfig, updateParams]
  )

  const setFilters = useCallback(
    (values: Partial<TListQueryValues<TFilters>>) => {
      updateParams((next, current) => {
        let resetOffset = false
        Object.entries(values).forEach(([key, value]) => {
          const parameter = filterConfig[key]
          if (!parameter) return
          const currentValue = readQueryParameter(current, parameter)
          if (
            parameter.resetsOffset !== false &&
            !equalQueryValues(parameter, currentValue, value)
          ) {
            resetOffset = true
          }
          writeQueryParameter(next, parameter, value)
        })
        if (resetOffset) writeQueryParameter(next, offsetConfig, 0)
      }, true)
    },
    [filterConfig, offsetConfig, updateParams]
  )

  const setFilter = useCallback(
    <K extends keyof TFilters>(key: K, value: TListQueryValues<TFilters>[K]) =>
      setFilters({ [key]: value } as Partial<TListQueryValues<TFilters>>),
    [setFilters]
  )

  const resetFilters = useCallback(() => {
    updateParams((next) => {
      Object.values(filterConfig).forEach(({ param }) => next.delete(param))
      writeQueryParameter(next, offsetConfig, 0)
    }, true)
  }, [filterConfig, offsetConfig, updateParams])

  return {
    search,
    offset,
    pageSize,
    filters,
    queryKey,
    setSearch,
    setOffset,
    setFilter,
    setFilters,
    resetFilters,
  }
}
