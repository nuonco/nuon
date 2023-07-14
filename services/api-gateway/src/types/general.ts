import { YogaInitialContext } from "graphql-yoga";

export type TDateTimeObject = {
  day: number;
  hours: number;
  minutes: number;
  month: number;
  nanos: number;
  seconds: number;
  timeZone?: string;
  utcOffset: { nanos: number; seconds: number };
  year: number;
};

export interface IGQLContext extends Partial<YogaInitialContext> {
  clients?: Record<string, any>;
  user?: Record<string, unknown>;
}

export type TResolverFn<A, O, I = undefined> = (
  info?: I,
  args?: A,
  ctx?: IGQLContext
) => Promise<Partial<O>> | Partial<O>;

export type TgRPCMessage = { toObject: () => Record<string, unknown> } & Record<
  string,
  unknown
>;
