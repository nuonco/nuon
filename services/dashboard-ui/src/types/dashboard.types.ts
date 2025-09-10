import type { ReactNode } from 'react'

export type TRouteParams<S extends string | number | symbol = string> = Record<
  S,
  string
>
export type TRouteSearchParams<S extends string | number | symbol = string> =
  Record<S, string>

export interface IPageProps<
  P extends string | number | symbol = string,
  S extends string | number | symbol = string,
> {
  params?: Promise<TRouteParams<P>>
  searchParams?: Promise<TRouteSearchParams<S>>
}

export interface ILayoutProps<
  P extends string | number | symbol = string,
  S extends string | number | symbol = string,
> {
  children: ReactNode
  params?: Promise<TRouteParams<P>>
  searchParams?: Promise<TRouteSearchParams<S>>
}

export interface IRouteProps extends IPageProps {}

export type TNavLink = {
  icon?: React.ReactNode
  path: string
  text: string
  isExternal?: boolean
}

export type TPaginationPageData = {
  hasNext: string
  offset: string
}

export type TPaginationParams = {
  offset?: number | string;
  limit?: number | string;
};

// fetch wrapper types
export type TAPIError = {
  description: string;
  error: string;
  user_error: boolean;
  meta?: any;
};

export type TAPIResponse<T> = {
  data: T | null;
  error: null | TAPIError;
  headers: Response["headers"];
  status: Response["status"];
};
