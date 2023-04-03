import type { Query, TResolverFn } from "../../types";

export const me: TResolverFn<undefined, Query["me"]> = (_, __, { user }) =>
  user;
