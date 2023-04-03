import { formatDateTime } from "../utils";

export function getNodeFields<T>(node): T {
  return {
    ...node,
    createdAt: formatDateTime(node.createdAt),
    updatedAt: formatDateTime(node.updatedAt),
  };
}
