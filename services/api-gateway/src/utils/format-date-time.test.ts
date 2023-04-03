import { TDateTimeObject } from "../types";
import { formatDateTime } from "./format-date-time";

test("formatDateTime should take a DateTime object & return a ISO date string", () => {
  const spec = formatDateTime({
    day: 31,
    hours: 8,
    minutes: 15,
    month: 12,
    seconds: 30,
    year: 1999,
  } as TDateTimeObject);

  expect(spec).toMatchInlineSnapshot(`"1999-12-31T08:15:30.000"`);
});

test("formatDateTime should take a DateTime object & return a ISO date string when having UTC offset", () => {
  const spec = formatDateTime({
    day: 31,
    hours: 8,
    minutes: 15,
    month: 12,
    seconds: 30,
    utcOffset: {
      nanos: 0,
      seconds: 0,
    },
    year: 1999,
  } as TDateTimeObject);

  expect(spec).toMatchInlineSnapshot(`"1999-12-31T08:15:30.000"`);
});

test("formatDateTime should return a current ISO date string when given undefined", () => {
  const spec = formatDateTime(undefined as TDateTimeObject);
  expect(spec).toBeDefined();
});
