import type { ReactNode } from 'react'

// TODO(nnnat): old types replace with types below
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
// -- end old types ---

// nextjs types
export type TParams<Keys extends string> = Promise<Record<Keys, string>>;

export type TRouteProps<Keys extends string, T = {}> = {
  params: TParams<Keys>;
} & T;

export type TPageProps<Keys extends string, T = {}> = {
  params: TParams<Keys>;
  searchParams: Promise<Record<string, string>>;
} & T;

export type TLayoutProps<Keys extends string, T = {}> = {
  children: ReactNode;
  params: TParams<Keys>;
} & T;

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
  headers: Record<string, string>;
  status: Response["status"];
};
