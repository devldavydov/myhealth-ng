import { type FormEvent, useEffect, useMemo, useState } from "react";
import { Link, useNavigate, useParams } from "react-router-dom";
import { ApiError, createFood, getFoodByKey, updateFood, type Food, type FoodData } from "../api";

type NutritionMode = "per100" | "portion";
type NumericField = "weight" | "cal100" | "prot100" | "fat100" | "carb100";
type FormState = {
  name: string;
  brand: string;
  cal100: string;
  prot100: string;
  fat100: string;
  carb100: string;
  comment: string;
  weight: string;
};
type FieldErrors = Partial<Record<keyof FormState, string>>;

const emptyForm: FormState = {
  name: "", brand: "", cal100: "", prot100: "", fat100: "", carb100: "", comment: "", weight: ""
};
const numericFields: NumericField[] = ["cal100", "prot100", "fat100", "carb100"];
const numberFormat = new Intl.NumberFormat("ru", { maximumFractionDigits: 2 });

function normalizeNumericInput(value: string): string {
  const digitsAndSeparators = value.replace(/[^0-9.,]/g, "");
  const separatorIndex = digitsAndSeparators.search(/[.,]/);
  if (separatorIndex === -1) return digitsAndSeparators;
  return digitsAndSeparators.slice(0, separatorIndex + 1) + digitsAndSeparators.slice(separatorIndex + 1).replace(/[.,]/g, "");
}

function parseNumber(value: string): number {
  const normalized = value.trim().replace(",", ".");
  return normalized === "" ? Number.NaN : Number(normalized);
}

function foodToForm(food: Food): FormState {
  return {
    name: food.name,
    brand: food.brand,
    cal100: String(food.cal100),
    prot100: String(food.prot100),
    fat100: String(food.fat100),
    carb100: String(food.carb100),
    comment: food.comment,
    weight: ""
  };
}

function validateForm(form: FormState, mode: NutritionMode, editing: boolean): FieldErrors {
  const errors: FieldErrors = {};
  if (!form.name.trim()) errors.name = "Введите название продукта";
  for (const field of numericFields) {
    const value = parseNumber(form[field]);
    if (form[field].trim() === "") errors[field] = "Укажите значение";
    else if (!Number.isFinite(value) || value < 0) errors[field] = "Введите неотрицательное число";
  }
  if (!editing && mode === "portion") {
    const weight = parseNumber(form.weight);
    if (form.weight.trim() === "") errors.weight = "Укажите вес продукта";
    else if (!Number.isFinite(weight) || weight <= 0) errors.weight = "Вес должен быть больше нуля";
  }
  return errors;
}

function requestFieldErrors(error: unknown): FieldErrors {
  if (!(error instanceof ApiError)) return {};
  return Object.fromEntries(
    Object.entries(error.details)
      .filter(([field, messages]) => field in emptyForm && messages.length > 0)
      .map(([field, messages]) => [field, messages[0]])
  ) as FieldErrors;
}

function FieldError({ message }: { message?: string }) {
  return message ? <span className="field-error" role="alert">{message}</span> : null;
}

export function FoodFormPage() {
  const { key } = useParams<{ key: string }>();
  const navigate = useNavigate();
  const editing = key !== undefined;
  const [form, setForm] = useState<FormState>(emptyForm);
  const [mode, setMode] = useState<NutritionMode>("per100");
  const [fieldErrors, setFieldErrors] = useState<FieldErrors>({});
  const [loading, setLoading] = useState(editing);
  const [saving, setSaving] = useState(false);
  const [loadError, setLoadError] = useState("");
  const [error, setError] = useState("");

  useEffect(() => {
    if (!key) return;
    let cancelled = false;
    setLoading(true);
    setLoadError("");
    void getFoodByKey(key)
      .then((food) => { if (!cancelled) setForm(foodToForm(food)); })
      .catch((requestError: unknown) => {
        if (!cancelled) setLoadError(requestError instanceof Error ? requestError.message : "Не удалось загрузить продукт");
      })
      .finally(() => { if (!cancelled) setLoading(false); });
    return () => { cancelled = true; };
  }, [key]);

  const converted = useMemo(() => {
    const multiplier = mode === "portion" ? 100 / parseNumber(form.weight) : 1;
    if (!Number.isFinite(multiplier) || multiplier <= 0) return null;
    const values = numericFields.map((field) => parseNumber(form[field]));
    if (values.some((value) => !Number.isFinite(value) || value < 0)) return null;
    return { cal100: values[0] * multiplier, prot100: values[1] * multiplier, fat100: values[2] * multiplier, carb100: values[3] * multiplier };
  }, [form, mode]);

  function setField(field: keyof FormState, value: string) {
    setForm((current) => ({ ...current, [field]: value }));
    setFieldErrors((current) => {
      if (!current[field]) return current;
      const next = { ...current };
      delete next[field];
      return next;
    });
  }

  function setNumericField(field: NumericField, value: string) {
    setField(field, normalizeNumericInput(value));
  }

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const validationErrors = validateForm(form, mode, editing);
    setError("");
    if (Object.keys(validationErrors).length > 0) {
      setFieldErrors(validationErrors);
      return;
    }
    if (!converted) return;

    const data: FoodData = { name: form.name, brand: form.brand, comment: form.comment, ...converted };
    setSaving(true);
    try {
      if (key) await updateFood(key, data);
      else await createFood(data);
      navigate("/food");
    } catch (requestError) {
      const errors = requestFieldErrors(requestError);
      if (Object.keys(errors).length > 0) setFieldErrors(errors);
      else setError(requestError instanceof Error ? requestError.message : "Не удалось сохранить продукт");
    } finally {
      setSaving(false);
    }
  }

  const nutritionLabel = mode === "portion" ? "на указанный вес" : "на 100 г";

  return (
    <div className="page food-form-page">
      <Link className="back-link" to="/food">← К продуктам</Link>
      <header className="form-page-header">
        <p className="eyebrow">{editing ? "Редактирование" : "Новый продукт"}</p>
        <h1>{editing ? "Изменить продукт" : "Добавить продукт"}</h1>
      </header>

      {loading ? <p className="muted">Загрузка продукта…</p> : loadError ? (
        <div className="error" role="alert">{loadError}</div>
      ) : (
        <section className="food-form-card">
          {error && <div className="error" role="alert">{error}</div>}
          <form noValidate onSubmit={handleSubmit}>
            <label>Название
              <input aria-invalid={Boolean(fieldErrors.name)} value={form.name} onChange={(event) => setField("name", event.target.value)} />
              <FieldError message={fieldErrors.name} />
            </label>
            <label>Бренд
              <input value={form.brand} onChange={(event) => setField("brand", event.target.value)} placeholder="Можно оставить пустым" />
            </label>
            {!editing && (
              <fieldset className="mode-switch">
                <legend>Как указаны КБЖУ</legend>
                <label><input name="nutrition-mode" type="radio" checked={mode === "per100"} onChange={() => { setMode("per100"); setFieldErrors((current) => ({ ...current, weight: undefined })); }} /> На 100 г</label>
                <label><input name="nutrition-mode" type="radio" checked={mode === "portion"} onChange={() => setMode("portion")} /> На другой вес</label>
              </fieldset>
            )}
            {mode === "portion" && !editing && (
              <label>Вес продукта, г
                <input aria-invalid={Boolean(fieldErrors.weight)} aria-label="Вес продукта, г" inputMode="decimal" pattern="[0-9]*[.,]?[0-9]*" value={form.weight} onChange={(event) => setNumericField("weight", event.target.value)} placeholder="Например, 250" />
                <FieldError message={fieldErrors.weight} />
              </label>
            )}
            <div className="nutrition-grid">
              <label>Ккал, {nutritionLabel}
                <input aria-invalid={Boolean(fieldErrors.cal100)} aria-label={`Ккал, ${nutritionLabel}`} inputMode="decimal" pattern="[0-9]*[.,]?[0-9]*" value={form.cal100} onChange={(event) => setNumericField("cal100", event.target.value)} />
                <FieldError message={fieldErrors.cal100} />
              </label>
              <label>Белки, г
                <input aria-invalid={Boolean(fieldErrors.prot100)} aria-label="Белки, г" inputMode="decimal" pattern="[0-9]*[.,]?[0-9]*" value={form.prot100} onChange={(event) => setNumericField("prot100", event.target.value)} />
                <FieldError message={fieldErrors.prot100} />
              </label>
              <label>Жиры, г
                <input aria-invalid={Boolean(fieldErrors.fat100)} aria-label="Жиры, г" inputMode="decimal" pattern="[0-9]*[.,]?[0-9]*" value={form.fat100} onChange={(event) => setNumericField("fat100", event.target.value)} />
                <FieldError message={fieldErrors.fat100} />
              </label>
              <label>Углеводы, г
                <input aria-invalid={Boolean(fieldErrors.carb100)} aria-label="Углеводы, г" inputMode="decimal" pattern="[0-9]*[.,]?[0-9]*" value={form.carb100} onChange={(event) => setNumericField("carb100", event.target.value)} />
                <FieldError message={fieldErrors.carb100} />
              </label>
            </div>
            {mode === "portion" && converted && (
              <div className="conversion-preview" aria-live="polite">
                На 100 г: {numberFormat.format(converted.cal100)} ккал · Б {numberFormat.format(converted.prot100)} · Ж {numberFormat.format(converted.fat100)} · У {numberFormat.format(converted.carb100)}
              </div>
            )}
            <label>Комментарий
              <textarea value={form.comment} onChange={(event) => setField("comment", event.target.value)} placeholder="Необязательно" />
            </label>
            <div className="form-actions">
              <button className="button primary" disabled={saving} type="submit">{saving ? "Сохраняем…" : editing ? "Сохранить изменения" : "Добавить"}</button>
              <Link className="button secondary" to="/food">Отмена</Link>
            </div>
          </form>
        </section>
      )}
    </div>
  );
}
