import type { ReactNode } from 'react'

export type TRouteParams<S extends string | number | symbol = string> = Record<
  S,
  string
>
export type TRouteSearchParams<S extends string | number | symbol = string> =
  Record<S, string | string[] | undefined>

export interface IPageProps<
  P extends string | number | symbol = string,
  S extends string | number | symbol = string,
> {
  params?: TRouteParams<P>
  searchParams?: TRouteSearchParams<S>
}

export interface ILayoutProps<
  P extends string | number | symbol = string,
  S extends string | number | symbol = string,
> extends IPageProps<P, S> {
  children: ReactNode
}

export interface IRouteProps extends IPageProps {}
