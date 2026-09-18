import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { afterEach, expect, it, vi } from "vitest";
import { SportActivityFormPage } from "./SportActivityFormPage";

afterEach(() => { cleanup(); vi.unstubAllGlobals(); });
function response(body: unknown) { return { ok: true, json: async () => body }; }

it("создаёт активность на отдельной странице через поисковый выбор спорта", async () => {
  const sport = { key: "турник", name: "Турник", unit: "шт", comment: "Подтягивания" };
  let saved: unknown;
  vi.stubGlobal("fetch", vi.fn().mockImplementation(async (input: string, init?: RequestInit) => {
    if (input.startsWith("/api/sport?")) return response({ data: [sport], pagination: { page: 1, pageSize: 50, total: 1, totalPages: 1 } });
    if (input === "/api/sport-activity" && init?.method === "POST") { saved=JSON.parse(String(init.body)); return response({ data: { ...(saved as object), sport } }); }
    return response({ data: [] });
  }));
  render(<MemoryRouter initialEntries={["/sport-activity/new"]}><Routes><Route path="/sport-activity/new" element={<SportActivityFormPage/>}/><Route path="/sport-activity" element={<h1>Активность</h1>}/></Routes></MemoryRouter>);
  const picker=screen.getByRole("combobox",{name:"Поиск вида спорта"});fireEvent.focus(picker);fireEvent.change(picker,{target:{value:"Тур"}});fireEvent.mouseDown(await screen.findByText("Турник"));fireEvent.click(screen.getByText("Турник"));
  fireEvent.change(screen.getByLabelText("Подход 1"),{target:{value:"5"}});fireEvent.click(screen.getByRole("button",{name:"Добавить подход"}));fireEvent.change(screen.getByLabelText("Подход 2"),{target:{value:"3,5"}});fireEvent.click(screen.getByRole("button",{name:"Сохранить"}));
  await waitFor(()=>expect(saved).toMatchObject({sportKey:"турник",sets:[5,3.5]}));expect(await screen.findByRole("heading",{name:"Активность"})).toBeInTheDocument();
});

it("загружает запись для редактирования на отдельной странице", async () => {
  const sport={key:"турник",name:"Турник",unit:"шт",comment:""};
  vi.stubGlobal("fetch",vi.fn().mockImplementation(async(input:string)=>input.startsWith("/api/sport-activity?")?response({data:[{dt:"2026-09-18",sport,sets:[5,3]}]}):response({data:[]})));
  render(<MemoryRouter initialEntries={["/sport-activity/2026-09-18/%D1%82%D1%83%D1%80%D0%BD%D0%B8%D0%BA/edit"]}><Routes><Route path="/sport-activity/:dt/:sportKey/edit" element={<SportActivityFormPage/>}/></Routes></MemoryRouter>);
  expect(await screen.findByRole("heading",{name:"Изменить активность"})).toBeInTheDocument();await waitFor(()=>expect(screen.getByLabelText("Подход 1")).toHaveValue("5"));expect(screen.getByLabelText("Дата")).toBeDisabled();expect(document.querySelector(`input[aria-label="Поиск вида спорта"]`)).toBeDisabled();
});
