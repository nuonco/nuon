const tint = (color, intensity) =>
  `color-mix(in srgb, ${color}, white ${intensity}%)`;

const shade = (color, intensity) =>
  `color-mix(in srgb, ${color}, black ${intensity}%)`;

// Generate the range of tints and shades required for tailwind.
export const generateColorRange = (color) => ({
  50: tint(color, 95),
  100: tint(color, 90),
  200: tint(color, 70),
  300: tint(color, 50),
  400: tint(color, 30),
  500: color,
  600: shade(color, 10),
  700: shade(color, 30),
  800: shade(color, 50),
  900: shade(color, 70),
  950: shade(color, 90),
});
