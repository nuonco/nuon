import type { AstroComponentFactory } from "astro/runtime/server/index.js";
import type { HTMLAttributes, ImageMetadata } from "astro/types";

export interface DayCard {
  /** A unique ID number that identifies a post. */
  id: string;

  /**  */
  date: any;

  /**  */
  title: string;

  link?: string;
  description?: string;

  // disable fields
  disableToday?: boolean;
  disableSeriesA?: boolean;
  disablePivot?: boolean;
}

export interface CodeLinkProps {
  repo?: string;
  title: string;
}
