import { echo } from "./echo";
import { ping } from "./ping";

export const generalResolvers = {
  Mutation: {
    echo,
  },
  Query: {
    ping,
  },
};
