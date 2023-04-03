import { YogaInitialContext } from "graphql-yoga";
import { TDateTimeObject } from "../types";

export interface IGQLContext extends Partial<YogaInitialContext> {
  clients?: Record<string, any>;
  user?: Record<string, unknown>;
}

export type TResolverFn<A, O, I = undefined> = (
  info?: I,
  args?: A,
  ctx?: IGQLContext
) => Promise<Partial<O>> | Partial<O>;

export type TNode = {
  createdAt: TDateTimeObject | string;
  id: string;
  updatedAt: TDateTimeObject | string;
};

export type TPageInfo = {
  endCursor: string;
  hasNextPage: boolean;
  hasPreviousPage: boolean;
  startCursor: string;
};

export type TEdge<T> = {
  cursor: string;
  node: T;
};

export type TConnection<T> = {
  edges: TEdge<T>[];
  pageInfo: TPageInfo;
  totalCount: number;
};

export type TConnectionOptions = {
  after?: string;
  before?: string;
  first?: number;
  last?: number;
};

export interface IConnectionResolver {
  options?: TConnectionOptions;
}
