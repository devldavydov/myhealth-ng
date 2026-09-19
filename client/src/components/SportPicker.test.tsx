import { cleanup, fireEvent, render, screen } from "@testing-library/react";
import { afterEach, expect, it, vi } from "vitest";
import { SportPicker } from "./SportPicker";

afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
});

it("рендерит меню поиска спорта под полем через portal", async () => {
  const sport = { key: "sport", name: "Приседания", unit: "кг", comment: "" };
  vi.stubGlobal("fetch", vi.fn().mockResolvedValue({
    ok: true,
    json: async () => ({
      data: [sport],
      pagination: { page: 1, pageSize: 50, total: 1, totalPages: 1 }
    })
  }));
  render(<SportPicker onError={() => undefined} onSelect={() => undefined} />);
  const picker = screen.getByRole("combobox", { name: "Поиск вида спорта" });
  fireEvent.focus(picker);
  fireEvent.change(picker, { target: { value: "Прис" } });
  await screen.findByText("Приседания");
  const portal = document.querySelector<HTMLElement>(".bundle-select__menu-portal");
  expect(portal).not.toBeNull();
  expect(portal?.parentElement).toBe(document.body);
  expect(portal).toHaveStyle({ position: "absolute" });
});
