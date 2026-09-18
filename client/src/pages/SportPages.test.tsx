import { cleanup, fireEvent, render, screen, waitFor, within } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, expect, it, vi } from "vitest";
import { buildSportChartData, formatSportChartLabel, formatSportChartTick, SportActivityOverviewPage, initialSportActivityRange } from "./SportActivityOverviewPage";

afterEach(() => { cleanup(); vi.unstubAllGlobals(); });

function response(body: unknown) { return { ok: true, json: async () => body }; }

it("задаёт период в шесть календарных месяцев", () => {
  expect(initialSportActivityRange(new Date(2026, 8, 18))).toEqual({ from: "2026-03-18", to: "2026-09-18" });
});

it("сохраняет отдельные категории для одного и двух столбцов графика", () => {
  const sport = { key: "run", name: "Бег", unit: "км", comment: "" };
  const activities = [
    { dt: "2026-09-18", sport, sets: [5] },
    { dt: "2026-09-17", sport, sets: [4, 2] }
  ];

  expect(buildSportChartData(activities.slice(0, 1), "run")).toEqual([
    { dt: "2026-09-18", set1: 5 }
  ]);
  expect(buildSportChartData(activities, "run")).toEqual([
    { dt: "2026-09-17", set1: 4, set2: 2 },
    { dt: "2026-09-18", set1: 5 }
  ]);
  expect(formatSportChartTick("2026-09-18")).toBe("18.09");
  expect(formatSportChartLabel("2026-09-18")).toContain("18 сентября 2026");
});

it("показывает отсортированный список и итоговую таблицу активности", async () => {
  const sportA = { key: "a", name: "Брусья", unit: "шт", comment: "" };
  const sportB = { key: "b", name: "Турник", unit: "шт", comment: "" };
  const activities = [
    { dt: "2026-09-18", sport: sportB, sets: [5, 3] },
    { dt: "2026-09-18", sport: sportA, sets: [8, 6] },
    { dt: "2026-09-17", sport: sportB, sets: [4] }
  ];
  vi.stubGlobal("fetch", vi.fn().mockImplementation(async (input: string) => {
    if (input.startsWith("/api/sport-activity?")) return response({ data: activities });
    if (input.startsWith("/api/sport?")) return response({ data: [sportA, sportB], pagination: { page: 1, pageSize: 50, total: 2, totalPages: 1 } });
    return response({ data: [] });
  }));
  render(<MemoryRouter><SportActivityOverviewPage /></MemoryRouter>);
  const list = await screen.findByRole("heading", { name: "Активности" });
  const rows = list.closest("section")!.querySelectorAll("article");
  expect(rows).toHaveLength(3);
  expect(rows[0]).toHaveTextContent("Брусья");
  expect(rows[1]).toHaveTextContent("Турник");
  expect(rows[1]).toHaveTextContent("Итого: 8 шт");
  fireEvent.click(screen.getByRole("tab", { name: "Статистика" }));
  const table = screen.getByRole("table");
  const tableRows = within(table).getAllByRole("row");
  expect(tableRows[1]).toHaveTextContent("Брусья14 шт");
  expect(tableRows[2]).toHaveTextContent("Турник12 шт");
  await waitFor(() => expect(screen.getByText("Выберите активность, чтобы построить график.")).toBeInTheDocument());
});
