import merge from "lodash.merge";
import { appResolvers } from "./app";
import { buildResolvers } from "./build";
import { componentResolvers } from "./component";
import { deployResolvers } from "./deploy";
import { deploymentResolvers } from "./deployment";
import { generalResolvers } from "./general";
import { githubResolvers } from "./github";
import { installResolvers } from "./install";
import { instanceResolvers } from "./instance";
import { orgResolvers } from "./org";
import { secretResolvers } from "./secret";
import { userResolvers } from "./user";

export const resolvers = merge(
  appResolvers,
  buildResolvers,
  componentResolvers,
  deployResolvers,
  deploymentResolvers,
  generalResolvers,
  githubResolvers,
  installResolvers,
  instanceResolvers,
  orgResolvers,
  secretResolvers,
  userResolvers
);
