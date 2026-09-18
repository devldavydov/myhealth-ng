import { expect, it } from "vitest";
import { categoricalYearMarkers, timeYearMarkers } from "./ChartYearLines";

it("показывает первый год даже для данных в пределах одного года", () => {
  expect(categoricalYearMarkers([{ dt: "2026-01-10" }, { dt: "2026-12-20" }])).toEqual([
    { value: "2026-01-10", year: 2026, separator: false }
  ]);
  expect(timeYearMarkers([
    { dt: "2026-01-10", timestamp: new Date(2026, 0, 10).getTime() },
    { dt: "2026-12-20", timestamp: new Date(2026, 11, 20).getTime() }
  ])).toEqual([
    { value: new Date(2026, 0, 10).getTime(), year: 2026, separator: false }
  ]);
});

it("ставит категориальные разделители в начале каждого нового года", () => {
  expect(categoricalYearMarkers([
    { dt: "2024-12-20" },
    { dt: "2025-02-10" },
    { dt: "2025-08-01" },
    { dt: "2026-03-15" }
  ])).toEqual([
    { value: "2024-12-20", year: 2024, separator: false },
    { value: "2025-02-10", year: 2025, separator: true },
    { value: "2026-03-15", year: 2026, separator: true }
  ]);
});

it("ставит временные разделители точно на первое января, включая годы без точек", () => {
  const first = new Date(2024, 11, 20).getTime();
  const last = new Date(2026, 2, 15).getTime();
  expect(timeYearMarkers([
    { dt: "2024-12-20", timestamp: first },
    { dt: "2026-03-15", timestamp: last }
  ])).toEqual([
    { value: first, year: 2024, separator: false },
    { value: new Date(2025, 0, 1).getTime(), year: 2025, separator: true },
    { value: new Date(2026, 0, 1).getTime(), year: 2026, separator: true }
  ]);
});
