import type { Mutation, MutationEchoArgs, TResolverFn } from "../../types";

export const echo: TResolverFn<MutationEchoArgs, Mutation["echo"]> = (
  _,
  { word }
) => word;
