import { TResolverFn } from "../../types";

export const me: TResolverFn<undefined, { id: string } | null> = (
  _,
  __,
  { user }
) => user;
