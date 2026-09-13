import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it, vi } from "vitest";
import { App } from "./App";

const user = { guid: "3f67c05f-7c9e-4cb5-b26a-f9ce5b065865", name: "Анна" };

function mockApi(measurements: unknown[] = []) {
  vi.stubGlobal("fetch", vi.fn().mockImplementation(async (input: string) => ({
    ok: true,
    json: async () => input === "/api/me" ? { data: user } : { data: measurements }
  })));
}

describe("SPA navigation", () => {
  afterEach(() => vi.unstubAllGlobals());

  it("показывает обзор и имя пользователя", async () => {
    mockApi();
    render(<MemoryRouter initialEntries={["/"]}><App /></MemoryRouter>);

    expect(screen.getByRole("heading", { name: "Добрый день!" })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "Добавить измерение" })).toHaveAttribute("href", "/measurements");
    expect(await screen.findByText("Анна")).toBeInTheDocument();
  });

  it("показывает страницу измерений", async () => {
    mockApi();
    render(<MemoryRouter initialEntries={["/measurements"]}><App /></MemoryRouter>);

    expect(screen.getByRole("heading", { name: "Измерения" })).toBeInTheDocument();
    expect(await screen.findByText("Измерений пока нет.")).toBeInTheDocument();
  });
});
