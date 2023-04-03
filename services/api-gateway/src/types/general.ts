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
