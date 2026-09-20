import { cleanup, fireEvent, render, screen, waitFor, within } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it, vi } from "vitest";
import { App } from "./App";
import { TimedNotification } from "./components/TimedNotification";
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
const meals = ["завтрак", "до обеда", "обед", "полдник", "до ужина", "ужин"] as const;
const emptyTotals = { weight: 0, cal: 0, protein: 0, fat: 0, carb: 0 };

function journalDay(withFood = false) {
  return {
    dt: "2026-09-16",
    zones: meals.map((meal) => meal === "завтрак" && withFood ? {
      meal,
      items: [{ food, weight: 200 }],
      totals: { weight: 200, cal: 240, protein: 36, fat: 10, carb: 6 }
    } : { meal, items: [], totals: emptyTotals }),
    totals: withFood ? { weight: 200, cal: 240, protein: 36, fat: 10, carb: 6 } : emptyTotals,
    macroPercent: withFood ? { protein: 69.2307, fat: 19.2307, carb: 11.5384 } : { protein: 0, fat: 0, carb: 0 }
  };
}

function response(body: unknown, ok = true) {
  if (body && typeof body === "object" && "data" in body && Array.isArray(body.data) && ("pagination" in body) === false) {
    body = {
      ...body,
      pagination: { page: 1, pageSize: 20, total: body.data.length, totalPages: body.data.length > 0 ? 1 : 0 }
    };
  }
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

  it("автоматически скрывает уведомление по круговому таймеру", async () => {
    vi.useFakeTimers();
    const onDismiss = vi.fn();
    render(<TimedNotification duration={1000} message="Продукт добавлен." onDismiss={onDismiss} />);

    expect(screen.getByRole("status")).toHaveTextContent("Продукт добавлен.");
    expect(screen.getByRole("img", { name: "До автоматического скрытия" })).toBeInTheDocument();
    expect(screen.getByRole("img", { name: "До автоматического скрытия" }).querySelector(".notification-timer-progress"))
      .toHaveStyle({ animationDuration: "1000ms" });

    await vi.advanceTimersByTimeAsync(999);
    expect(onDismiss).not.toHaveBeenCalled();
    await vi.advanceTimersByTimeAsync(1);
    expect(onDismiss).toHaveBeenCalledOnce();
  });

  it("показывает дашборд на корневом адресе без отдельного пункта меню", async () => {
    vi.useFakeTimers({ toFake: ["Date"] });
    vi.setSystemTime(new Date(2026, 8, 20, 12));
    const fetchMock = vi.fn().mockImplementation(async (input: string) => {
      if (input === "/api/me") return response({ data: user });
      if (input === "/api/dashboard?from=2026-08-20&to=2026-09-20") {
        return response({ data: { calories: { days: [], average: null }, weightChange: null, activities: [] } });
      }
      return response({ data: [] });
    });
    vi.stubGlobal("fetch", fetchMock);
    render(<MemoryRouter initialEntries={["/"]}><App /></MemoryRouter>);

    expect(await screen.findByRole("heading", { name: "Главная" })).toBeInTheDocument();
    expect(screen.getByLabelText("От")).toHaveValue("2026-08-20");
    expect(screen.getByLabelText("До")).toHaveValue("2026-09-20");
    await waitFor(() => expect(fetchMock).toHaveBeenCalledWith("/api/dashboard?from=2026-08-20&to=2026-09-20", undefined));
    expect(screen.getByRole("link", { name: "M+ MyHealth" })).toHaveAttribute("href", "/");
    expect(within(screen.getByRole("navigation", { name: "Основная навигация" })).queryByText("Главная")).not.toBeInTheDocument();
    expect(await screen.findByText("Анна")).toBeInTheDocument();
  });

  it("перенаправляет неизвестный адрес на главную", async () => {
    vi.stubGlobal("fetch", vi.fn().mockImplementation(async (input: string) =>
      input === "/api/me"
        ? response({ data: user })
        : response({ data: { calories: { days: [], average: null }, weightChange: null, activities: [] } })
    ));
    render(<MemoryRouter initialEntries={["/missing/page"]}><App /></MemoryRouter>);

    expect(await screen.findByRole("heading", { name: "Главная" })).toBeInTheDocument();
  });

  it("раскрывает и закрывает меню разделов", async () => {
    vi.stubGlobal("fetch", vi.fn().mockImplementation(async (input: string) =>
      input === "/api/me" ? response({ data: user }) : response({ data: [] })
    ));
    render(<MemoryRouter initialEntries={["/food"]}><App /></MemoryRouter>);

    const toggle = screen.getByLabelText("Выбрать раздел");
    expect(toggle).toHaveTextContent("Еда");
    expect(toggle.querySelector("svg.navigation-icon")).toBeInTheDocument();
    expect(toggle.querySelector("svg.mobile-nav-chevron")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "Вес" }).querySelector("svg.navigation-icon")).toBeInTheDocument();
    expect(toggle).toHaveAttribute("aria-expanded", "false");
    fireEvent.click(toggle);
    expect(toggle).toHaveAttribute("aria-expanded", "true");

    fireEvent.keyDown(document, { key: "Escape" });
    expect(toggle).toHaveAttribute("aria-expanded", "false");
    fireEvent.click(toggle);
    fireEvent.pointerDown(document.body);
    expect(toggle).toHaveAttribute("aria-expanded", "false");

    fireEvent.click(toggle);
    fireEvent.click(screen.getByRole("link", { name: "Вес" }));
    expect(toggle).toHaveAttribute("aria-expanded", "false");
  });

  it("показывает пустые настройки и сохраняет числовой лимит", async () => {
    let savedBody: { defaultDailyCalorieLimit: number } | undefined;
    const fetchMock = vi.fn().mockImplementation(async (input: string, init?: RequestInit) => {
      if (input === "/api/me") return response({ data: user });
      if (input === "/api/settings" && init?.method === "PUT") {
        savedBody = JSON.parse(String(init.body)) as { defaultDailyCalorieLimit: number };
        return response({ data: savedBody });
      }
      if (input === "/api/settings") return response({ data: { defaultDailyCalorieLimit: null } });
      if (input.startsWith("/api/active-calories")) return response({ data: null });
      return response({ data: [] });
    });
    vi.stubGlobal("fetch", fetchMock);
    render(<MemoryRouter initialEntries={["/settings"]}><App /></MemoryRouter>);

    expect(await screen.findByRole("heading", { name: "Настройки" })).toBeInTheDocument();
    const limit = await screen.findByLabelText("Лимит ккал в день по умолчанию");
    expect(limit).toHaveValue("");
    fireEvent.change(limit, { target: { value: "около 2500 ккал" } });
    expect(limit).toHaveValue("2500");
    fireEvent.click(screen.getByRole("button", { name: "Сохранить" }));

    await waitFor(() => expect(savedBody).toEqual({ defaultDailyCalorieLimit: 2500 }));
    expect(await screen.findByText("Настройки сохранены.")).toBeInTheDocument();
  });

  it("показывает ошибку лимита под полем", async () => {
    vi.stubGlobal("fetch", vi.fn().mockImplementation(async (input: string) =>
      input === "/api/me" ? response({ data: user }) : response({ data: { defaultDailyCalorieLimit: null } })
    ));
    render(<MemoryRouter initialEntries={["/settings"]}><App /></MemoryRouter>);
    const limit = await screen.findByLabelText("Лимит ккал в день по умолчанию");
    fireEvent.change(limit, { target: { value: "10001" } });
    fireEvent.click(screen.getByRole("button", { name: "Сохранить" }));

    expect(await screen.findByText("Введите целое число от 1 до 10000")).toBeInTheDocument();
    expect(limit).toHaveAttribute("aria-invalid", "true");
  });

  it("показывает журнал, массовые проценты БЖУ и удаляет продукт после подтверждения", async () => {
    let current = journalDay(true);
    const itemURL = `/api/journal/2026-09-16/${encodeURIComponent("завтрак")}/${food.key}`;
    const fetchMock = vi.fn().mockImplementation(async (input: string, init?: RequestInit) => {
      if (input === "/api/me") return response({ data: user });
      if (input === "/api/settings") return response({ data: { defaultDailyCalorieLimit: 200 } });
      if (input.startsWith("/api/active-calories")) return response({ data: null });
      if (input === itemURL && init?.method === "DELETE") {
        current = journalDay(false);
        return response({ data: current });
      }
      if (input === "/api/journal?dt=2026-09-16") return response({ data: current });
      return response({ data: [] });
    });
    vi.stubGlobal("fetch", fetchMock);
    render(<MemoryRouter initialEntries={["/journal?dt=2026-09-16"]}><App /></MemoryRouter>);

    expect(await screen.findByRole("heading", { name: "Журнал" })).toBeInTheDocument();
    expect(screen.getByLabelText("Дата журнала")).toHaveValue("2026-09-16");
    expect(await screen.findByText("69,2%")).toBeInTheDocument();
    expect(screen.getByText("120% · 240 из 200 ккал")).toBeInTheDocument();
    expect(screen.getByLabelText("Прогресс дневного лимита")).toHaveClass("exceeded");
    fireEvent.click(screen.getByRole("button", { name: "Удалить Творог из завтрак" }));
    const dialog = screen.getByRole("dialog", { name: "Удалить продукт из журнала?" });
    fireEvent.click(within(dialog).getByRole("button", { name: "Удалить" }));

    await waitFor(() => expect(fetchMock).toHaveBeenCalledWith(itemURL, { method: "DELETE" }));
    expect(await screen.findByText("Продукт «Творог» удалён.")).toBeInTheDocument();
  });

  it("возвращает журнал на сегодняшнюю дату", async () => {
    const today = (() => {
      const value = new Date();
      const year = String(value.getFullYear()).padStart(4, "0");
      const month = String(value.getMonth() + 1).padStart(2, "0");
      const day = String(value.getDate()).padStart(2, "0");
      return `${year}-${month}-${day}`;
    })();
    const fetchMock = vi.fn().mockImplementation(async (input: string) => {
      if (input === "/api/me") return response({ data: user });
      if (input === "/api/settings") return response({ data: { defaultDailyCalorieLimit: null } });
      if (input.startsWith("/api/active-calories")) return response({ data: null });
      if (input.startsWith("/api/journal?dt=")) return response({ data: { ...journalDay(false), dt: input.slice(input.indexOf("=") + 1) } });
      return response({ data: [] });
    });
    vi.stubGlobal("fetch", fetchMock);
    render(<MemoryRouter initialEntries={["/journal?dt=2026-09-15"]}><App /></MemoryRouter>);
    await screen.findByRole("heading", { name: "завтрак" });
    fireEvent.click(screen.getByRole("button", { name: "Сегодня" }));

    await waitFor(() => expect(screen.getByLabelText("Дата журнала")).toHaveValue(today));
    expect(fetchMock).toHaveBeenCalledWith(`/api/journal?dt=${today}`, undefined);
    expect(screen.getByRole("button", { name: "Сегодня" })).toBeDisabled();
  });

  it("добавляет продукт с весом в выбранную зону журнала", async () => {
    let savedBody: Record<string, unknown> | undefined;
    const fetchMock = vi.fn().mockImplementation(async (input: string, init?: RequestInit) => {
      if (input === "/api/me") return response({ data: user });
      if (input === "/api/settings") return response({ data: { defaultDailyCalorieLimit: null } });
      if (input.startsWith("/api/active-calories")) return response({ data: null });
      if (input.startsWith("/api/food")) return response({ data: [food] });
      if (input.startsWith("/api/bundle")) return response({ data: [] });
      if (input === "/api/journal" && init?.method === "POST") {
        savedBody = JSON.parse(String(init.body)) as Record<string, unknown>;
        return response({ data: journalDay(true) });
      }
      if (input === "/api/journal?dt=2026-09-16") return response({ data: journalDay(false) });
      return response({ data: [] });
    });
    vi.stubGlobal("fetch", fetchMock);
    render(<MemoryRouter initialEntries={["/journal?dt=2026-09-16"]}><App /></MemoryRouter>);
    const heading = await screen.findByRole("heading", { name: "завтрак" });
    const zone = heading.closest("section")!;
    fireEvent.click(within(zone).getByRole("button", { name: "Добавить" }));
    const picker = within(zone).getByRole("combobox", { name: "Поиск еды или бандла" });
    fireEvent.focus(picker);
    fireEvent.change(picker, { target: { value: "Твор" } });
    const option = await screen.findByText("Творог");
    fireEvent.mouseDown(option);
    fireEvent.click(option);
    const weightInput = await within(zone).findByLabelText("Вес продукта Творог");
    expect(weightInput).toHaveFocus();
    fireEvent.change(weightInput, { target: { value: "порция 150,5 г" } });
    fireEvent.click(within(zone).getAllByRole("button", { name: "Добавить" }).at(-1)!);

    await waitFor(() => expect(savedBody).toEqual({
      dt: "2026-09-16",
      meal: "завтрак",
      items: [{ foodKey: food.key, weight: 150.5 }]
    }));
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
    expect(fetchMock).toHaveBeenCalledWith("/api/food?q=%D0%A2%D0%B2%D0%BE%D1%80%D0%BE%D0%B3+5%25&page=1&pageSize=20", undefined);
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
    let current = { ...food, cal100: 120.06, prot100: 18.44, fat100: 5.55, carb100: 3.04 };
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
    expect(screen.getByLabelText("Ккал, на 100 г")).toHaveValue("120,1");
    expect(screen.getByLabelText("Белки, г")).toHaveValue("18,4");
    expect(screen.getByLabelText("Жиры, г")).toHaveValue("5,6");
    expect(screen.getByLabelText("Углеводы, г")).toHaveValue("3");
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
      if (input.startsWith("/api/bundle?")) return response({ data: deleted ? [] : [bundleSummary] });
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
