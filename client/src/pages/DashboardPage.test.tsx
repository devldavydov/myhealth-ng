import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it, vi } from "vitest";
import { calorieBarFill, DashboardPage, initialDashboardRange } from "./DashboardPage";

function response(data: unknown) {
  return { ok: true, json: async () => ({ data }) };
}

describe("страница аналитики", () => {
  afterEach(() => {
    cleanup();
    vi.unstubAllGlobals();
    vi.useRealTimers();
  });

  it("выбирает предыдущий календарный месяц с корректировкой конца месяца", () => {
    expect(initialDashboardRange(new Date(2026, 2, 31, 12))).toEqual({ from: "2026-02-28", to: "2026-03-31" });
    expect(initialDashboardRange(new Date(2024, 2, 31, 12))).toEqual({ from: "2024-02-29", to: "2024-03-31" });
  });

  it("назначает цвет столбца по знаку баланса", () => {
    expect(calorieBarFill(100)).toBe("#2b9185");
    expect(calorieBarFill(0)).toBe("#2b9185");
    expect(calorieBarFill(-100)).toBe("#c9535e");
  });

  it("загружает и показывает все показатели", async () => {
    vi.useFakeTimers({ toFake: ["Date"] });
    vi.setSystemTime(new Date(2026, 2, 31, 12));
    const fetchMock = vi.fn().mockResolvedValue(response({
      calories: {
        days: [
          { dt: "2026-03-29", balance: 100 },
          { dt: "2026-03-30", balance: -100 }
        ],
        average: 0
      },
      weightChange: -1.5,
      activities: [{ sportKey: "walk", name: "Ходьба", count: 2, total: 20, unit: "км" }]
    }));
    vi.stubGlobal("fetch", fetchMock);

    render(<MemoryRouter><DashboardPage /></MemoryRouter>);

    expect(await screen.findByRole("heading", { name: "Аналитика" })).toBeInTheDocument();
    expect(screen.getByLabelText("От")).toHaveValue("2026-02-28");
    expect(screen.getByLabelText("До")).toHaveValue("2026-03-31");
    await waitFor(() => expect(fetchMock).toHaveBeenCalledWith("/api/dashboard?from=2026-02-28&to=2026-03-31", undefined));
    expect(screen.getByText("В среднем баланс калорий равен нулю")).toBeInTheDocument();
    expect(screen.getByText("−1,5 кг")).toBeInTheDocument();
    expect(screen.getByText("Ходьба")).toBeInTheDocument();
    expect(screen.getByText("2 раз")).toBeInTheDocument();
    expect(screen.getByText("20 км")).toBeInTheDocument();
  });

  it("просит задать дефолт и валидирует диапазон до запроса", async () => {
    const fetchMock = vi.fn().mockResolvedValue(response({ calories: null, weightChange: null, activities: [] }));
    vi.stubGlobal("fetch", fetchMock);
    render(<MemoryRouter><DashboardPage /></MemoryRouter>);

    expect(await screen.findByText(/задайте/i)).toBeInTheDocument();
    expect(screen.getByRole("link", { name: /дефолтный лимит/i })).toHaveAttribute("href", "/settings");
    fireEvent.change(screen.getByLabelText("От"), { target: { value: "2026-10-01" } });
    fireEvent.change(screen.getByLabelText("До"), { target: { value: "2026-09-01" } });
    fireEvent.click(screen.getByRole("button", { name: "Показать" }));

    expect(screen.getByText("Дата «от» должна быть не позже даты «до»")).toBeInTheDocument();
    expect(fetchMock).toHaveBeenCalledTimes(1);
  });
});
