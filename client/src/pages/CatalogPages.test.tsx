import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it, vi } from "vitest";
import { App } from "../App";

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

function response(body: unknown) {
  return { ok: true, json: async () => body };
}

describe("страничные каталоги", () => {
  afterEach(() => {
    cleanup();
    vi.unstubAllGlobals();
  });

  it("читает фильтр и страницу из URL и меняет размер страницы", async () => {
    const fetchMock = vi.fn().mockImplementation(async (input: string) => {
      if (input === "/api/me") return response({ data: user });
      if (input.startsWith("/api/food?")) {
        const params = new URL(input, "http://localhost").searchParams;
        const page = Number(params.get("page"));
        const pageSize = Number(params.get("pageSize"));
        return response({
          data: [food],
          pagination: { page, pageSize, total: 25, totalPages: Math.ceil(25 / pageSize) }
        });
      }
      return response({ data: null });
    });
    vi.stubGlobal("fetch", fetchMock);

    render(<MemoryRouter initialEntries={["/food?q=%D0%A2%D0%B2%D0%BE%D1%80&page=2&pageSize=10"]}><App /></MemoryRouter>);

    expect(await screen.findByRole("heading", { name: "Творог" })).toBeInTheDocument();
    expect(screen.getByLabelText("Поиск продуктов")).toHaveValue("Твор");
    expect(screen.getByLabelText("Найдено продуктов: 25")).toBeInTheDocument();
    expect(screen.getByText("Страница 2 из 3")).toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: "Следующая страница" }));
    await waitFor(() => expect(fetchMock).toHaveBeenCalledWith(
      "/api/food?q=%D0%A2%D0%B2%D0%BE%D1%80&page=3&pageSize=10",
      undefined
    ));

    fireEvent.change(screen.getByLabelText("Количество продуктов на странице"), { target: { value: "50" } });
    await waitFor(() => expect(fetchMock).toHaveBeenCalledWith(
      "/api/food?q=%D0%A2%D0%B2%D0%BE%D1%80&page=1&pageSize=50",
      undefined
    ));
    expect(await screen.findByText("Страница 1 из 1")).toBeInTheDocument();
  });
});
