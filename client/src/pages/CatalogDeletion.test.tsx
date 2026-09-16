import { cleanup, fireEvent, render, screen, waitFor, within } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it, vi } from "vitest";
import { App } from "../App";

const user = { guid: "3f67c05f-7c9e-4cb5-b26a-f9ce5b065865", name: "Анна" };
const food = {
  key: "c39dbf56-73ce-4630-b86c-a08c313eab75",
  name: "Последний продукт",
  brand: "",
  cal100: 100,
  prot100: 1,
  fat100: 2,
  carb100: 3,
  comment: ""
};

function response(body: unknown) {
  return { ok: true, json: async () => body };
}

describe("удаление с последней страницы каталога", () => {
  afterEach(() => {
    cleanup();
    vi.unstubAllGlobals();
  });

  it("переходит на последнюю существующую страницу", async () => {
    let deleted = false;
    const fetchMock = vi.fn().mockImplementation(async (input: string, init?: RequestInit) => {
      if (input === "/api/me") return response({ data: user });
      if (input === `/api/food/${food.key}` && init?.method === "DELETE") {
        deleted = true;
        return { ok: true };
      }
      if (input.startsWith("/api/food?")) {
        const params = new URL(input, "http://localhost").searchParams;
        const page = Number(params.get("page"));
        if (!deleted) {
          return response({ data: [food], pagination: { page, pageSize: 10, total: 11, totalPages: 2 } });
        }
        return response({ data: [], pagination: { page, pageSize: 10, total: 10, totalPages: 1 } });
      }
      return response({ data: null });
    });
    vi.stubGlobal("fetch", fetchMock);

    render(<MemoryRouter initialEntries={["/food?page=2&pageSize=10"]}><App /></MemoryRouter>);
    expect(await screen.findByRole("heading", { name: food.name })).toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: "Удалить" }));
    fireEvent.click(within(screen.getByRole("dialog", { name: "Удалить продукт?" })).getByRole("button", { name: "Удалить" }));

    await waitFor(() => expect(fetchMock).toHaveBeenCalledWith("/api/food?page=1&pageSize=10", undefined));
    expect(await screen.findByText("Страница 1 из 1")).toBeInTheDocument();
  });
});
