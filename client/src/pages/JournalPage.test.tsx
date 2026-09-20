import { cleanup, fireEvent, render, screen, waitFor, within } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it, vi } from "vitest";
import { JournalPage } from "./JournalPage";

const dt = "2026-09-16";
const emptyTotals = { weight: 0, cal: 0, protein: 0, fat: 0, carb: 0 };
const day = {
  dt,
  zones: ["завтрак", "до обеда", "обед", "полдник", "до ужина", "ужин"].map((meal) => ({
    meal,
    items: [],
    totals: emptyTotals
  })),
  totals: { ...emptyTotals, cal: 240 },
  macroPercent: { protein: 0, fat: 0, carb: 0 }
};

function response(body: unknown, ok = true) {
  return { ok, json: async () => body };
}

describe("активные калории в журнале", () => {
  afterEach(() => {
    cleanup();
    vi.unstubAllGlobals();
  });

  it("использует дневное значение вместо дефолта и сохраняет числовой ввод", async () => {
    let savedBody: unknown;
    const fetchMock = vi.fn().mockImplementation(async (input: string, init?: RequestInit) => {
      if (input === `/api/journal?dt=${dt}`) return response({ data: day });
      if (input === "/api/settings") return response({ data: { defaultDailyCalorieLimit: 200 } });
      if (input === `/api/active-calories?dt=${dt}`) return response({ data: { dt, value: 300 } });
      if (input === "/api/active-calories" && init?.method === "POST") {
        savedBody = JSON.parse(String(init.body));
        return response({ data: savedBody });
      }
      return response({ data: [] });
    });
    vi.stubGlobal("fetch", fetchMock);
    render(<MemoryRouter initialEntries={[`/journal?dt=${dt}`]}><JournalPage /></MemoryRouter>);

    const input = await screen.findByLabelText("Активные калории за день");
    expect(input).toHaveValue("300");
    expect(screen.getByText("80% · 240 из 300 ккал")).toBeInTheDocument();
    expect(screen.getByText("Осталось 60 ккал")).toHaveClass("remaining");
    expect(screen.getByText("Лимит за выбранный день")).toBeInTheDocument();

    fireEvent.change(input, { target: { value: "около 2750,5 ккал" } });
    expect(input).toHaveValue("2750,5");
    fireEvent.click(screen.getByRole("button", { name: "Сохранить" }));

    await waitFor(() => expect(savedBody).toEqual({ dt, value: 2750.5 }));
    expect(await screen.findByText("Активные калории за день сохранены.")).toBeInTheDocument();
  });

  it("после подтверждённого удаления возвращается к дефолтному лимиту", async () => {
    const fetchMock = vi.fn().mockImplementation(async (input: string, init?: RequestInit) => {
      if (input === `/api/journal?dt=${dt}`) return response({ data: day });
      if (input === "/api/settings") return response({ data: { defaultDailyCalorieLimit: 200 } });
      if (input === `/api/active-calories?dt=${dt}`) return response({ data: { dt, value: 300 } });
      if (input === `/api/active-calories/${dt}` && init?.method === "DELETE") return { ok: true };
      return response({ data: [] });
    });
    vi.stubGlobal("fetch", fetchMock);
    render(<MemoryRouter initialEntries={[`/journal?dt=${dt}`]}><JournalPage /></MemoryRouter>);

    fireEvent.click(await screen.findByRole("button", { name: "Убрать значение" }));
    const dialog = screen.getByRole("dialog", { name: "Убрать активные калории за день?" });
    fireEvent.click(within(dialog).getByRole("button", { name: "Убрать" }));

    await waitFor(() => expect(fetchMock).toHaveBeenCalledWith(`/api/active-calories/${dt}`, { method: "DELETE" }));
    expect(await screen.findByText("120% · 240 из 200 ккал")).toBeInTheDocument();
    expect(screen.getByText("Перерасход 40 ккал")).toHaveClass("over");
    expect(screen.getByText("Дефолтный дневной лимит")).toBeInTheDocument();
    expect(screen.getByLabelText("Прогресс дневного лимита")).toHaveClass("exceeded");
  });

  it("показывает полевую ошибку и подсказку, когда ни одного лимита нет", async () => {
    vi.stubGlobal("fetch", vi.fn().mockImplementation(async (input: string) => {
      if (input === `/api/journal?dt=${dt}`) return response({ data: day });
      if (input === "/api/settings") return response({ data: { defaultDailyCalorieLimit: null } });
      if (input === `/api/active-calories?dt=${dt}`) return response({ data: null });
      return response({ data: [] });
    }));
    render(<MemoryRouter initialEntries={[`/journal?dt=${dt}`]}><JournalPage /></MemoryRouter>);

    await screen.findByLabelText("Активные калории за день");
    expect(screen.getByText(/Задайте активные калории за день/)).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "Сохранить" }));
    expect(await screen.findByText("Укажите активные калории")).toBeInTheDocument();
  });

  it("редактирует вес продукта прямо в журнале", async () => {
    const food = {
      key: "творог", name: "Творог", brand: "Ферма",
      cal100: 120, prot100: 18, fat100: 5, carb100: 3, comment: ""
    };
    const journal = {
      ...day,
      zones: day.zones.map((zone) => zone.meal === "завтрак"
        ? { ...zone, items: [{ food, weight: 123.456 }] }
        : zone)
    };
    let savedBody: unknown;
    vi.stubGlobal("fetch", vi.fn().mockImplementation(async (input: string, init?: RequestInit) => {
      if (input === `/api/journal?dt=${dt}`) return response({ data: journal });
      if (input === "/api/settings") return response({ data: { defaultDailyCalorieLimit: 300 } });
      if (input === `/api/active-calories?dt=${dt}`) return response({ data: null });
      if (input === "/api/journal" && init?.method === "POST") {
        savedBody = JSON.parse(String(init.body));
        return response({ data: journal });
      }
      return response({ data: [] });
    }));
    render(<MemoryRouter initialEntries={[`/journal?dt=${dt}`]}><JournalPage /></MemoryRouter>);

    expect(await screen.findByRole("link", { name: "Редактировать продукт Творог" }))
      .toHaveAttribute("href", "/food/%D1%82%D0%B2%D0%BE%D1%80%D0%BE%D0%B3/edit");
    fireEvent.click(await screen.findByRole("button", { name: "Редактировать Творог в завтрак" }));
    const weightInput = screen.getByLabelText("Вес продукта Творог");
    expect(weightInput).toHaveValue("123,5");
    expect(weightInput).toHaveFocus();
    fireEvent.change(weightInput, { target: { value: "150,5" } });
    fireEvent.click(screen.getByRole("button", { name: "Сохранить вес продукта Творог" }));

    await waitFor(() => expect(savedBody).toEqual({
      dt, meal: "завтрак", items: [{ foodKey: "творог", weight: 150.5 }]
    }));
    expect(await screen.findByText("Вес продукта «Творог» изменён.")).toBeInTheDocument();
  });
});
