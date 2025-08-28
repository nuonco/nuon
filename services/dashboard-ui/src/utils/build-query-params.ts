export const buildQueryParams = (params: Record<string, any>): string => {
  const filtered = Object.entries(params).filter(
    ([_, v]) => v !== undefined && v !== null,
  );
  if (filtered.length === 0) return "";
  const query = new URLSearchParams(Object.fromEntries(filtered)).toString();
  return `?${query}`;
};
