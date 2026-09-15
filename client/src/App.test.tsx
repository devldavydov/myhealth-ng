import { cleanup, fireEvent, render, screen, waitFor, within } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it, vi } from "vitest";
import { App } from "./App";
import { mergeBundleItems } from "./pages/BundleFormPage";
import { initialWeightRange } from "./pages/WeightPage";

const user = { guid: "3f67c05f-7c9e-4cb5-b26a-f9ce5b065865", name: "Анна" };
const food = {
  key: "c39dbf56-73ce-4630-b86c-a08c313eab75",
  name: "Творог",
  brand: "Ферма",
  cal100: 120,
  prot100: 18,
  fat100: 5,
  carb100: 3,
  comment: "5%"
};
const bundleKey = "8c2cf7aa-44f4-49a4-aef0-e9087088c980";
const bundleSummary = {
  key: bundleKey,
  name: "Завтрак",
  itemCount: 1,
  totals: { weight: 200, cal: 240, protein: 36, fat: 10, carb: 6 }
};
const bundle = {
  ...bundleSummary,
  items: [{ food, weight: 200 }]
};

function response(body: unknown, ok = true) {
  return { ok, json: async () => body };
}

describe("MyHealth SPA", () => {
  afterEach(() => {
    cleanup();
    vi.unstubAllGlobals();
	vi.useRealTimers();
  });

  it("суммирует пересекающиеся продукты при разворачивании бандла", () => {
    const merged = mergeBundleItems(
      [{ food, weight: "40" }],
      [{ food, weight: 200 }]
    );
    expect(merged).toHaveLength(1);
    expect(merged[0].weight).toBe("240");
  });

  it("перенаправляет на список продуктов и показывает пользователя", async () => {
    vi.stubGlobal("fetch", vi.fn().mockImplementation(async (input: string) =>
      input === "/api/me" ? response({ data: user }) : response({ data: [] })
    ));
    render(<MemoryRouter initialEntries={["/"]}><App /></MemoryRouter>);

    expect(await screen.findByRole("heading", { name: "Продукты" })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "Еда" })).toHaveAttribute("href", "/food");
    expect(screen.getByRole("link", { name: "Вес" })).toHaveAttribute("href", "/weight");
    expect(screen.getByRole("link", { name: "Добавить продукт" })).toHaveAttribute("href", "/food/new");
    expect(await screen.findByText("Анна")).toBeInTheDocument();
    expect(await screen.findByText("Продуктов пока нет. Добавьте первый.")).toBeInTheDocument();
  });

  it("показывает историю веса за последние шесть календарных месяцев", async () => {
    expect(initialWeightRange(new Date(2026, 8, 15, 10))).toEqual({ from: "2026-03-15", to: "2026-09-15" });
    const entry = { dt: "2026-09-15", value: 82.4 };
    const fetchMock = vi.fn().mockImplementation(async (input: string) =>
      input === "/api/me" ? response({ data: user }) : response({ data: [entry] })
    );
    vi.stubGlobal("fetch", fetchMock);
    render(<MemoryRouter initialEntries={["/weight"]}><App /></MemoryRouter>);

    expect(await screen.findByRole("heading", { name: "Вес" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "Динамика веса" })).toBeInTheDocument();
    await waitFor(() => expect(fetchMock).toHaveBeenCalledWith(expect.stringMatching(/^\/api\/weight\?from=\d{4}-\d{2}-\d{2}&to=\d{4}-\d{2}-\d{2}$/), undefined));
    expect(await screen.findByText("15 сентября 2026 г.")).toBeInTheDocument();
    expect(screen.getByLabelText("Найдено измерений: 1")).toBeInTheDocument();
  });

  it("сохраняет числовое значение веса и обновляет список", async () => {
    const entry = { dt: "2026-09-15", value: 82.4 };
    let savedBody: typeof entry | undefined;
    const fetchMock = vi.fn().mockImplementation(async (input: string, init?: RequestInit) => {
      if (input === "/api/me") return response({ data: user });
      if (input === "/api/weight" && init?.method === "POST") {
        savedBody = JSON.parse(String(init.body)) as typeof entry;
        return response({ data: savedBody });
      }
      return response({ data: savedBody ? [savedBody] : [] });
    });
    vi.stubGlobal("fetch", fetchMock);
    render(<MemoryRouter initialEntries={["/weight"]}><App /></MemoryRouter>);
    await screen.findByText("Добавьте первое измерение за выбранный период.");

    fireEvent.change(screen.getByLabelText("Дата"), { target: { value: entry.dt } });
    fireEvent.change(screen.getByLabelText("Вес, кг"), { target: { value: "вес 82,4 кг" } });
    expect(screen.getByLabelText("Вес, кг")).toHaveValue("82,4");
    fireEvent.click(screen.getByRole("button", { name: "Сохранить" }));

    await waitFor(() => expect(savedBody).toEqual(entry));
    expect(await screen.findByLabelText("Найдено измерений: 1")).toBeInTheDocument();
  });

  it("проверяет диапазон и спрашивает подтверждение перед удалением веса", async () => {
    const entry = { dt: "2026-09-15", value: 82.4 };
    let deleted = false;
    const itemURL = `/api/weight/${entry.dt}`;
    const fetchMock = vi.fn().mockImplementation(async (input: string, init?: RequestInit) => {
      if (input === "/api/me") return response({ data: user });
      if (input === itemURL && init?.method === "DELETE") {
        deleted = true;
        return { ok: true };
      }
      return response({ data: deleted ? [] : [entry] });
    });
    vi.stubGlobal("fetch", fetchMock);
    render(<MemoryRouter initialEntries={["/weight"]}><App /></MemoryRouter>);
    await screen.findByText("15 сентября 2026 г.");

    fireEvent.change(screen.getByLabelText("От"), { target: { value: "2026-09-16" } });
    fireEvent.change(screen.getByLabelText("До"), { target: { value: "2026-09-15" } });
    fireEvent.click(screen.getByRole("button", { name: "Показать" }));
    expect(await screen.findByText("Дата «от» должна быть не позже даты «до»")).toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: "Удалить" }));
    let dialog = screen.getByRole("dialog", { name: "Удалить измерение?" });
    fireEvent.click(within(dialog).getByRole("button", { name: "Отмена" }));
    expect(fetchMock).not.toHaveBeenCalledWith(itemURL, { method: "DELETE" });

    fireEvent.click(screen.getByRole("button", { name: "Удалить" }));
    dialog = screen.getByRole("dialog", { name: "Удалить измерение?" });
    fireEvent.click(within(dialog).getByRole("button", { name: "Удалить" }));
    await waitFor(() => expect(fetchMock).toHaveBeenCalledWith(itemURL, { method: "DELETE" }));
    expect(await screen.findByText("Добавьте первое измерение за выбранный период.")).toBeInTheDocument();
  });

  it("ищет продукты через API", async () => {
    const fetchMock = vi.fn().mockImplementation(async (input: string) =>
      input === "/api/me" ? response({ data: user }) : response({ data: input.includes("?q=") ? [food] : [] })
    );
    vi.stubGlobal("fetch", fetchMock);
    render(<MemoryRouter initialEntries={["/food"]}><App /></MemoryRouter>);
    await screen.findByText("Продуктов пока нет. Добавьте первый.");

    fireEvent.change(screen.getByLabelText("Поиск продуктов"), { target: { value: "Творог 5%" } });
    fireEvent.click(screen.getByRole("button", { name: "Найти" }));

    expect(await screen.findByRole("heading", { name: "Творог" })).toBeInTheDocument();
    expect(fetchMock).toHaveBeenCalledWith("/api/food?q=%D0%A2%D0%B2%D0%BE%D1%80%D0%BE%D0%B3%205%25", undefined);
  });

  it("добавляет продукт на отдельной странице и пересчитывает порцию", async () => {
    let createdBody: Record<string, unknown> | undefined;
    const fetchMock = vi.fn().mockImplementation(async (input: string, init?: RequestInit) => {
      if (input === "/api/me") return response({ data: user });
      if (input === "/api/food" && init?.method === "POST") {
        createdBody = JSON.parse(String(init.body)) as Record<string, unknown>;
        return response({ data: { ...food, ...createdBody } });
      }
      return response({ data: [] });
    });
    vi.stubGlobal("fetch", fetchMock);
    render(<MemoryRouter initialEntries={["/food/new"]}><App /></MemoryRouter>);
    expect(await screen.findByRole("heading", { name: "Добавить продукт" })).toBeInTheDocument();

    fireEvent.change(screen.getByLabelText("Название"), { target: { value: "Йогурт" } });
    fireEvent.click(screen.getByLabelText("На другой вес"));
    fireEvent.change(screen.getByLabelText("Вес продукта, г"), { target: { value: "250" } });
    fireEvent.change(screen.getByLabelText("Ккал, на указанный вес"), { target: { value: "200" } });
    fireEvent.change(screen.getByLabelText("Белки, г"), { target: { value: "10" } });
    fireEvent.change(screen.getByLabelText("Жиры, г"), { target: { value: "5" } });
    fireEvent.change(screen.getByLabelText("Углеводы, г"), { target: { value: "20" } });
    fireEvent.click(screen.getByRole("button", { name: "Добавить" }));

    await waitFor(() => expect(createdBody).toBeDefined());
    expect(createdBody).toMatchObject({ name: "Йогурт", cal100: 80, prot100: 4, fat100: 2, carb100: 8 });
    expect(await screen.findByRole("heading", { name: "Продукты" })).toBeInTheDocument();
  });

  it("не принимает текст в КБЖУ и показывает ошибку под полем", async () => {
    const fetchMock = vi.fn().mockImplementation(async (input: string) =>
      input === "/api/me" ? response({ data: user }) : response({ data: [] })
    );
    vi.stubGlobal("fetch", fetchMock);
    render(<MemoryRouter initialEntries={["/food/new"]}><App /></MemoryRouter>);
    await screen.findByRole("heading", { name: "Добавить продукт" });

    fireEvent.change(screen.getByLabelText("Название"), { target: { value: "Йогурт" } });
    fireEvent.change(screen.getByLabelText("Ккал, на 100 г"), { target: { value: "двести" } });
    fireEvent.change(screen.getByLabelText("Белки, г"), { target: { value: "10" } });
    fireEvent.change(screen.getByLabelText("Жиры, г"), { target: { value: "5" } });
    fireEvent.change(screen.getByLabelText("Углеводы, г"), { target: { value: "20" } });
    expect(screen.getByLabelText("Ккал, на 100 г")).toHaveValue("");

    fireEvent.click(screen.getByRole("button", { name: "Добавить" }));
    expect(await screen.findByText("Укажите значение")).toBeInTheDocument();
    expect(screen.getByLabelText("Ккал, на 100 г")).toHaveAttribute("aria-invalid", "true");
    expect(fetchMock.mock.calls.some(([input, init]) => input === "/api/food" && init?.method === "POST")).toBe(false);
  });

  it("загружает и редактирует продукт на отдельной странице", async () => {
    let current = food;
    const itemURL = `/api/food/${food.key}`;
    const fetchMock = vi.fn().mockImplementation(async (input: string, init?: RequestInit) => {
      if (input === "/api/me") return response({ data: user });
      if (input === itemURL && !init) return response({ data: current });
      if (input === itemURL && init?.method === "PUT") {
        current = { ...current, ...(JSON.parse(String(init.body)) as typeof food) };
        return response({ data: current });
      }
      return response({ data: [current] });
    });
    vi.stubGlobal("fetch", fetchMock);
    render(<MemoryRouter initialEntries={[`/food/${food.key}/edit`]}><App /></MemoryRouter>);

    expect(await screen.findByRole("heading", { name: "Изменить продукт" })).toBeInTheDocument();
    await waitFor(() => expect(screen.getByLabelText("Название")).toHaveValue("Творог"));
    fireEvent.change(screen.getByLabelText("Название"), { target: { value: "Творог мягкий" } });
    fireEvent.click(screen.getByRole("button", { name: "Сохранить изменения" }));

    expect(await screen.findByRole("heading", { name: "Творог мягкий" })).toBeInTheDocument();
    expect(fetchMock).toHaveBeenCalledWith(itemURL, expect.objectContaining({ method: "PUT" }));
  });

  it("спрашивает подтверждение перед удалением продукта", async () => {
    let deleted = false;
    const itemURL = `/api/food/${food.key}`;
    const fetchMock = vi.fn().mockImplementation(async (input: string, init?: RequestInit) => {
      if (input === "/api/me") return response({ data: user });
      if (input === itemURL && init?.method === "DELETE") {
        deleted = true;
        return { ok: true };
      }
      return response({ data: deleted ? [] : [food] });
    });
    vi.stubGlobal("fetch", fetchMock);
    render(<MemoryRouter initialEntries={["/food"]}><App /></MemoryRouter>);
    await screen.findByRole("heading", { name: "Творог" });

    expect(screen.getByRole("link", { name: "Редактировать" })).toHaveAttribute("href", `/food/${food.key}/edit`);
    fireEvent.click(screen.getByRole("button", { name: "Удалить" }));
    let dialog = screen.getByRole("dialog", { name: "Удалить продукт?" });
    expect(within(dialog).getByText("Продукт «Творог» будет удалён без возможности восстановления.")).toBeInTheDocument();
    expect(fetchMock).not.toHaveBeenCalledWith(itemURL, { method: "DELETE" });

    fireEvent.click(within(dialog).getByRole("button", { name: "Отмена" }));
    expect(screen.queryByRole("dialog", { name: "Удалить продукт?" })).not.toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "Удалить" }));
    dialog = screen.getByRole("dialog", { name: "Удалить продукт?" });
    fireEvent.click(within(dialog).getByRole("button", { name: "Удалить" }));

    await waitFor(() => expect(fetchMock).toHaveBeenCalledWith(itemURL, { method: "DELETE" }));
    expect(await screen.findByText("Продуктов пока нет. Добавьте первый.")).toBeInTheDocument();
  });

  it("показывает список бандлов и подтверждает удаление", async () => {
    let deleted = false;
    const itemURL = "/api/bundle/" + bundleKey;
    const fetchMock = vi.fn().mockImplementation(async (input: string, init?: RequestInit) => {
      if (input === "/api/me") return response({ data: user });
      if (input === itemURL && init?.method === "DELETE") {
        deleted = true;
        return { ok: true };
      }
      if (input === "/api/bundle") return response({ data: deleted ? [] : [bundleSummary] });
      return response({ data: [] });
    });
    vi.stubGlobal("fetch", fetchMock);
    render(<MemoryRouter initialEntries={["/bundle"]}><App /></MemoryRouter>);

    expect(await screen.findByRole("heading", { name: "Завтрак" })).toBeInTheDocument();
    expect(screen.getByText("1 продуктов · 200 г")).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "Удалить" }));
    const dialog = screen.getByRole("dialog", { name: "Удалить бандл?" });
    fireEvent.click(within(dialog).getByRole("button", { name: "Удалить" }));

    await waitFor(() => expect(fetchMock).toHaveBeenCalledWith(itemURL, { method: "DELETE" }));
    expect(await screen.findByText("Бандлов пока нет. Добавьте первый.")).toBeInTheDocument();
  });

  it("находит вложенный бандл и сохраняет его как плоский список еды", async () => {
    let createdBody: Record<string, unknown> | undefined;
    const itemURL = "/api/bundle/" + bundleKey;
    const fetchMock = vi.fn().mockImplementation(async (input: string, init?: RequestInit) => {
      if (input === "/api/me") return response({ data: user });
      if (input.startsWith("/api/food")) return response({ data: [food] });
      if (input === itemURL) return response({ data: bundle });
      if (input.startsWith("/api/bundle?")) return response({ data: [bundleSummary] });
      if (input === "/api/bundle" && init?.method === "POST") {
        createdBody = JSON.parse(String(init.body)) as Record<string, unknown>;
        return response({ data: bundle });
      }
      if (input === "/api/bundle") return response({ data: [bundleSummary] });
      return response({ data: [] });
    });
    vi.stubGlobal("fetch", fetchMock);
    render(<MemoryRouter initialEntries={["/bundle/new"]}><App /></MemoryRouter>);

    expect(await screen.findByRole("heading", { name: "Добавить бандл" })).toBeInTheDocument();
    fireEvent.change(screen.getByLabelText("Название"), { target: { value: "Рабочий завтрак" } });
    const picker = screen.getByRole("combobox", { name: "Поиск еды или бандла" });
    fireEvent.focus(picker);
    fireEvent.change(picker, { target: { value: "Зав" } });
    const option = await screen.findByText("Завтрак");
    fireEvent.mouseDown(option);
    fireEvent.click(option);

    await waitFor(() => expect(screen.getByLabelText("Вес продукта Творог")).toHaveValue("200"));
    expect(screen.getByText("Бандл «Завтрак» развёрнут в продукты.")).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "Добавить" }));

    await waitFor(() => expect(createdBody).toEqual({
      name: "Рабочий завтрак",
      items: [{ foodKey: food.key, weight: 200 }]
    }));
  });
});
