import { useEffect, useId, useRef, type FormEvent } from "react";
import { createPortal } from "react-dom";
import type { MealType } from "../api";

const mealTypes: MealType[] = ["завтрак", "до обеда", "обед", "полдник", "до ужина", "ужин"];

type CopyMealDialogProps = {
  open: boolean;
  busy: boolean;
  error: string;
  sourceDate: string;
  sourceMeal: MealType;
  targetDate: string;
  targetMeal: MealType;
  onCancel: () => void;
  onSourceDateChange: (value: string) => void;
  onSourceMealChange: (value: MealType) => void;
  onSubmit: (event: FormEvent<HTMLFormElement>) => void;
};

export function CopyMealDialog({
  open,
  busy,
  error,
  sourceDate,
  sourceMeal,
  targetDate,
  targetMeal,
  onCancel,
  onSourceDateChange,
  onSourceMealChange,
  onSubmit
}: CopyMealDialogProps) {
  const titleId = useId();
  const descriptionId = useId();
  const dialogRef = useRef<HTMLDivElement>(null);
  const dateInputRef = useRef<HTMLInputElement>(null);
  const busyRef = useRef(busy);
  const onCancelRef = useRef(onCancel);

  busyRef.current = busy;
  onCancelRef.current = onCancel;

  useEffect(() => {
    if (!open) return;

    const previousFocus = document.activeElement instanceof HTMLElement ? document.activeElement : null;
    const previousOverflow = document.body.style.overflow;
    document.body.style.overflow = "hidden";
    const focusFrame = window.requestAnimationFrame(() => dateInputRef.current?.focus());

    function handleKeyDown(event: KeyboardEvent) {
      if (event.key === "Escape" && !busyRef.current) {
        event.preventDefault();
        onCancelRef.current();
        return;
      }
      if (event.key !== "Tab") return;

      const focusable = dialogRef.current?.querySelectorAll<HTMLElement>(
        "button:not(:disabled), input:not(:disabled), select:not(:disabled)"
      );
      if (!focusable?.length) return;
      const first = focusable[0];
      const last = focusable[focusable.length - 1];
      if (event.shiftKey && document.activeElement === first) {
        event.preventDefault();
        last.focus();
      } else if (!event.shiftKey && document.activeElement === last) {
        event.preventDefault();
        first.focus();
      }
    }

    document.addEventListener("keydown", handleKeyDown);
    return () => {
      window.cancelAnimationFrame(focusFrame);
      document.removeEventListener("keydown", handleKeyDown);
      document.body.style.overflow = previousOverflow;
      previousFocus?.focus();
    };
  }, [open]);

  if (!open) return null;

  return createPortal(
    <div className="dialog-backdrop" onMouseDown={(event) => {
      if (event.target === event.currentTarget && !busy) onCancel();
    }}>
      <div
        aria-describedby={descriptionId}
        aria-labelledby={titleId}
        aria-modal="true"
        aria-busy={busy}
        className="dialog-card"
        ref={dialogRef}
        role="dialog"
      >
        <p className="eyebrow">Копирование</p>
        <h2 id={titleId}>Копировать в «{targetMeal}»</h2>
        <p className="dialog-message" id={descriptionId}>Целевая дата: {targetDate}</p>
        <form className="copy-meal-form" noValidate onSubmit={onSubmit}>
          <label>Дата источника
            <input
              aria-invalid={!sourceDate}
              onChange={(event) => onSourceDateChange(event.target.value)}
              ref={dateInputRef}
              required
              type="date"
              value={sourceDate}
            />
          </label>
          <label>Приём пищи
            <select onChange={(event) => onSourceMealChange(event.target.value as MealType)} value={sourceMeal}>
              {mealTypes.map((meal) => <option key={meal} value={meal}>{meal}</option>)}
            </select>
          </label>
          {error && <div className="error copy-meal-error" role="alert">{error}</div>}
          <div className="dialog-actions">
            <button className="button secondary" disabled={busy} onClick={onCancel} type="button">Отмена</button>
            <button className="button primary" disabled={busy} type="submit">{busy ? "Копируем…" : "Копировать"}</button>
          </div>
        </form>
      </div>
    </div>,
    document.body
  );
}
