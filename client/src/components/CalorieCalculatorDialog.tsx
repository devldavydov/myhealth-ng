import { useEffect, useId, useRef, useState, type FormEvent } from "react";
import { createPortal } from "react-dom";

type Sex = "male" | "female";

type CalorieCalculation = {
  maintenance: number;
  loss: number;
  gain: number;
};

type CalorieCalculatorDialogProps = {
  open: boolean;
  onCancel: () => void;
  onSelect: (value: number) => void;
};

type FieldErrors = Partial<Record<"age" | "height" | "weight", string>>;

const activityLevels = [
  { value: 1.2, label: "Минимальная", description: "Сидячий образ жизни, тренировок почти нет." },
  { value: 1.375, label: "Лёгкая", description: "Лёгкие тренировки или активные прогулки 1–3 раза в неделю." },
  { value: 1.55, label: "Средняя", description: "Тренировки средней интенсивности 3–5 раз в неделю." },
  { value: 1.725, label: "Высокая", description: "Интенсивные тренировки 6–7 раз в неделю." },
  { value: 1.9, label: "Очень высокая", description: "Тяжёлая физическая работа или интенсивные тренировки каждый день." }
] as const;

const calorieFormat = new Intl.NumberFormat("ru-RU");

export function calculateDailyCalories(
  sex: Sex,
  age: number,
  height: number,
  weight: number,
  activity: number
): CalorieCalculation {
  const resting = 10 * weight + 6.25 * height - 5 * age + (sex === "male" ? 5 : -161);
  const dailyExpenditure = resting * activity;

  return {
    maintenance: Math.round(dailyExpenditure),
    loss: Math.round(dailyExpenditure * 0.85),
    gain: Math.round(dailyExpenditure * 1.1)
  };
}

function normalizeDecimal(value: string): string {
  const cleaned = value.replace(/[^0-9.,]/g, "");
  const separator = cleaned.search(/[.,]/);
  return separator < 0 ? cleaned : cleaned.slice(0, separator + 1) + cleaned.slice(separator + 1).replace(/[.,]/g, "");
}

function parseDecimal(value: string): number {
  return Number(value.replace(",", "."));
}

export function CalorieCalculatorDialog({ open, onCancel, onSelect }: CalorieCalculatorDialogProps) {
  const titleId = useId();
  const descriptionId = useId();
  const dialogRef = useRef<HTMLDivElement>(null);
  const sexInputRef = useRef<HTMLSelectElement>(null);
  const onCancelRef = useRef(onCancel);
  const [sex, setSex] = useState<Sex>("female");
  const [age, setAge] = useState("");
  const [height, setHeight] = useState("");
  const [weight, setWeight] = useState("");
  const [activity, setActivity] = useState(1.2);
  const [errors, setErrors] = useState<FieldErrors>({});
  const [calculation, setCalculation] = useState<CalorieCalculation | null>(null);

  onCancelRef.current = onCancel;

  useEffect(() => {
    if (!open) return;

    const previousFocus = document.activeElement instanceof HTMLElement ? document.activeElement : null;
    const previousOverflow = document.body.style.overflow;
    document.body.style.overflow = "hidden";
    const focusFrame = window.requestAnimationFrame(() => sexInputRef.current?.focus());

    function handleKeyDown(event: KeyboardEvent) {
      if (event.key === "Escape") {
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

  const activityDescription = activityLevels.find((level) => level.value === activity)?.description ?? "";

  function invalidateCalculation() {
    setCalculation(null);
    setErrors({});
  }

  function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const ageValue = Number(age);
    const heightValue = Number(height);
    const weightValue = parseDecimal(weight);
    const nextErrors: FieldErrors = {};

    if (!Number.isInteger(ageValue) || ageValue < 18 || ageValue > 120) {
      nextErrors.age = "Укажите возраст от 18 до 120 лет";
    }
    if (!Number.isInteger(heightValue) || heightValue < 100 || heightValue > 250) {
      nextErrors.height = "Укажите рост от 100 до 250 см";
    }
    if (!Number.isFinite(weightValue) || weightValue < 30 || weightValue > 300) {
      nextErrors.weight = "Укажите вес от 30 до 300 кг";
    }

    setErrors(nextErrors);
    if (Object.keys(nextErrors).length > 0) {
      setCalculation(null);
      return;
    }

    setCalculation(calculateDailyCalories(sex, ageValue, heightValue, weightValue, activity));
  }

  return createPortal(
    <div className="dialog-backdrop" onMouseDown={(event) => {
      if (event.target === event.currentTarget) onCancel();
    }}>
      <div
        aria-describedby={descriptionId}
        aria-labelledby={titleId}
        aria-modal="true"
        className="dialog-card calorie-calculator-card"
        ref={dialogRef}
        role="dialog"
      >
        <p className="eyebrow">Расчёт дневной нормы</p>
        <h2 id={titleId}>Калькулятор калорий</h2>
        <p className="dialog-message" id={descriptionId}>
          Оценка для взрослых по формуле Миффлина — Сан Жеора. Рост и возраст нужны для расчёта базового обмена.
        </p>

        <form className="calorie-calculator-form" noValidate onSubmit={handleSubmit}>
          <div className="calorie-calculator-fields">
            <label>Пол
              <select
                onChange={(event) => { setSex(event.target.value as Sex); invalidateCalculation(); }}
                ref={sexInputRef}
                value={sex}
              >
                <option value="female">Женский</option>
                <option value="male">Мужской</option>
              </select>
            </label>
            <label>Возраст, лет
              <input
                aria-invalid={Boolean(errors.age)}
                inputMode="numeric"
                onChange={(event) => { setAge(event.target.value.replace(/\D/g, "")); invalidateCalculation(); }}
                pattern="[0-9]*"
                placeholder="Например, 35"
                value={age}
              />
              {errors.age && <span className="field-error" role="alert">{errors.age}</span>}
            </label>
            <label>Рост, см
              <input
                aria-invalid={Boolean(errors.height)}
                inputMode="numeric"
                onChange={(event) => { setHeight(event.target.value.replace(/\D/g, "")); invalidateCalculation(); }}
                pattern="[0-9]*"
                placeholder="Например, 170"
                value={height}
              />
              {errors.height && <span className="field-error" role="alert">{errors.height}</span>}
            </label>
            <label>Вес, кг
              <input
                aria-invalid={Boolean(errors.weight)}
                inputMode="decimal"
                onChange={(event) => { setWeight(normalizeDecimal(event.target.value)); invalidateCalculation(); }}
                pattern="[0-9]*[.,]?[0-9]*"
                placeholder="Например, 70"
                value={weight}
              />
              {errors.weight && <span className="field-error" role="alert">{errors.weight}</span>}
            </label>
            <label className="calorie-activity-field">Активность
              <select
                aria-label="Активность"
                onChange={(event) => { setActivity(Number(event.target.value)); invalidateCalculation(); }}
                value={activity}
              >
                {activityLevels.map((level) => (
                  <option key={level.value} value={level.value}>{level.label} · ×{level.value}</option>
                ))}
              </select>
              <span className="form-hint">{activityDescription}</span>
            </label>
          </div>

          <button className="button primary calorie-calculate-button" type="submit">Рассчитать</button>

          {calculation && (
            <div aria-live="polite" className="calorie-results">
              <p>Выберите значение, чтобы подставить его в поле лимита:</p>
              <div className="calorie-result-grid">
                <button
                  aria-label={`Выбрать поддержание веса: ${calculation.maintenance} ккал`}
                  className="calorie-result"
                  onClick={() => onSelect(calculation.maintenance)}
                  type="button"
                >
                  <span>По формуле</span>
                  <strong>{calorieFormat.format(calculation.maintenance)} ккал</strong>
                  <small>Поддержание веса</small>
                </button>
                <button
                  aria-label={`Выбрать снижение веса: ${calculation.loss} ккал`}
                  className="calorie-result loss"
                  onClick={() => onSelect(calculation.loss)}
                  type="button"
                >
                  <span>Для снижения</span>
                  <strong>{calorieFormat.format(calculation.loss)} ккал</strong>
                  <small>Умеренный дефицит 15%</small>
                </button>
                <button
                  aria-label={`Выбрать набор веса: ${calculation.gain} ккал`}
                  className="calorie-result gain"
                  onClick={() => onSelect(calculation.gain)}
                  type="button"
                >
                  <span>Для набора</span>
                  <strong>{calorieFormat.format(calculation.gain)} ккал</strong>
                  <small>Умеренный профицит 10%</small>
                </button>
              </div>
            </div>
          )}

          <p className="calorie-disclaimer">Результат ориентировочный и не заменяет рекомендации врача или диетолога.</p>
          <div className="dialog-actions">
            <button className="button secondary" onClick={onCancel} type="button">Закрыть</button>
          </div>
        </form>
      </div>
    </div>,
    document.body
  );
}
