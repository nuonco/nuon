import { DateTime } from "luxon";
import { TDateTimeObject } from "../types";

export function formatDateTime(dateObject: TDateTimeObject): string {
  if (dateObject != null) {
    const utcOffset = dateObject.utcOffset;
    if (utcOffset != null) {
      // if the UTC offset is zero, the timezone is UTC
      if (utcOffset.nanos == 0 && utcOffset.seconds == 0) {
        return DateTime.utc(
          dateObject?.year,
          dateObject?.month,
          dateObject?.day,
          dateObject?.hours,
          dateObject?.minutes,
          dateObject?.seconds
        )
          .toLocal()
          .toISO({ includeOffset: false });
      }
    }
  }

  return DateTime.fromObject({
    day: dateObject?.day,
    hours: dateObject?.hours,
    minutes: dateObject?.minutes,
    month: dateObject?.month,
    seconds: dateObject?.seconds,
    year: dateObject?.year,
  }).toISO({ includeOffset: false });
}
