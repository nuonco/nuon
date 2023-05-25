import { DateTime } from "luxon";
import { TDateTimeObject } from "../types";

export function formatDateTime(dateObject: TDateTimeObject): string {
  if (!dateObject) {
    return DateTime.utc().toISO();
  }
  return DateTime.utc(
    dateObject.year,
    dateObject.month,
    dateObject.day,
    dateObject.hours,
    dateObject.minutes,
    dateObject.seconds
  ).toISO();
}
