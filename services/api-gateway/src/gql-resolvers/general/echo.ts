import { TResolverFn } from "../../types";

export const echo: TResolverFn<{ word: string }, string> = (_, { word }) =>
  word;
