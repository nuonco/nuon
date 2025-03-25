
type TOrgID = 'org-id' | string 
export type TParams<T extends TOrgID> = Record<T, string>;
export type TRouteRes<T extends TOrgID> = { params: TParams<T>; }
