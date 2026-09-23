import { cleanup, fireEvent, render, screen, waitFor, within } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { CalorieCalculatorDialog, calculateDailyCalories } from "./CalorieCalculatorDialog";

describe("CalorieCalculatorDialog", () => {
  afterEach(cleanup);

  it("рассчитывает поддержание, снижение и набор по формуле Миффлина — Сан Жеора", () => {
    expect(calculateDailyCalories("male", 35, 180, 80, 1.2)).toEqual({
      maintenance: 2106,
      loss: 1790,
      gain: 2317
    });
    expect(calculateDailyCalories("female", 30, 165, 70, 1.55)).toEqual({
      maintenance: 2201,
      loss: 1871,
      gain: 2422
    });
  });

  it("передаёт выбранный результат без самостоятельного сохранения", async () => {
    const onSelect = vi.fn();
    render(<CalorieCalculatorDialog open onCancel={vi.fn()} onSelect={onSelect} />);
    const dialog = screen.getByRole("dialog", { name: "Калькулятор калорий" });

    await waitFor(() => expect(within(dialog).getByLabelText("Пол")).toHaveFocus());
    fireEvent.change(within(dialog).getByLabelText("Возраст, лет"), { target: { value: "30" } });
    fireEvent.change(within(dialog).getByLabelText("Рост, см"), { target: { value: "165" } });
    fireEvent.change(within(dialog).getByLabelText("Вес, кг"), { target: { value: "70" } });
    fireEvent.change(within(dialog).getByLabelText("Активность"), { target: { value: "1.55" } });
    fireEvent.click(within(dialog).getByRole("button", { name: "Рассчитать" }));

    expect(within(dialog).getByText("2 201 ккал")).toBeInTheDocument();
    expect(within(dialog).getByText("1 871 ккал")).toBeInTheDocument();
    expect(within(dialog).getByText("2 422 ккал")).toBeInTheDocument();
    fireEvent.click(within(dialog).getByRole("button", { name: "Выбрать снижение веса: 1871 ккал" }));
    expect(onSelect).toHaveBeenCalledWith(1871);
  });

  it("проверяет обязательные параметры расчёта", () => {
    render(<CalorieCalculatorDialog open onCancel={vi.fn()} onSelect={vi.fn()} />);
    const dialog = screen.getByRole("dialog", { name: "Калькулятор калорий" });
    fireEvent.click(within(dialog).getByRole("button", { name: "Рассчитать" }));

    expect(within(dialog).getByText("Укажите возраст от 18 до 120 лет")).toBeInTheDocument();
    expect(within(dialog).getByText("Укажите рост от 100 до 250 см")).toBeInTheDocument();
    expect(within(dialog).getByText("Укажите вес от 30 до 300 кг")).toBeInTheDocument();
  });
});
