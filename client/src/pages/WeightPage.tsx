import { type FormEvent, useCallback, useEffect, useMemo, useState } from "react";
import {
  CartesianGrid,
  Line,
  LineChart,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis
} from "recharts";
import { ApiError, deleteWeight, getWeight, saveWeight, type WeightEntry } from "../api";
import { ConfirmDialog } from "../components/ConfirmDialog";

type Range = { from: string; to: string };
type RangeErrors = Partial<Record<keyof Range, string>>;
type WeightFormErrors = Partial<Record<"dt" | "value", string>>;

const weightFormat = new Intl.NumberFormat("ru", { maximumFractionDigits: 2 });
const dateFormat = new Intl.DateTimeFormat("ru", { day: "numeric", month: "long", year: "numeric" });
const shortDateFormat = new Intl.DateTimeFormat("ru", { day: "2-digit", month: "2-digit" });

function localISODate(value: Date): string {
  const year = String(value.getFullYear()).padStart(4, "0");
  const month = String(value.getMonth() + 1).padStart(2, "0");
  const day = String(value.getDate()).padStart(2, "0");
  return `${year}-${month}-${day}`;
}

export function initialWeightRange(today = new Date()): Range {
  const targetMonth = new Date(today.getFullYear(), today.getMonth() - 6, 1);
  const lastTargetDay = new Date(targetMonth.getFullYear(), targetMonth.getMonth() + 1, 0).getDate();
  const from = new Date(targetMonth.getFullYear(), targetMonth.getMonth(), Math.min(today.getDate(), lastTargetDay));
  return { from: localISODate(from), to: localISODate(today) };
}

function parseLocalDate(value: string): Date {
  return new Date(`${value}T00:00:00`);
}

function formatDate(value: string): string {
  return dateFormat.format(parseLocalDate(value));
}

function normalizeNumericInput(value: string): string {
  const digitsAndSeparators = value.replace(/[^0-9.,]/g, "");
  const separatorIndex = digitsAndSeparators.search(/[.,]/);
  if (separatorIndex === -1) return digitsAndSeparators;
  return digitsAndSeparators.slice(0, separatorIndex + 1) + digitsAndSeparators.slice(separatorIndex + 1).replace(/[.,]/g, "");
}

function requestFormErrors(error: unknown): WeightFormErrors {
  if (!(error instanceof ApiError)) return {};
  const errors: WeightFormErrors = {};
  if (error.details.dt?.[0]) errors.dt = error.details.dt[0];
  if (error.details.value?.[0]) errors.value = error.details.value[0];
  return errors;
}

function FieldError({ message }: { message?: string }) {
  return message ? <span className="field-error" role="alert">{message}</span> : null;
}

export function WeightPage() {
  const [defaultRange] = useState(initialWeightRange);
  const [range, setRange] = useState<Range>(defaultRange);
  const [activeRange, setActiveRange] = useState<Range>(defaultRange);
  const [rangeErrors, setRangeErrors] = useState<RangeErrors>({});
  const [items, setItems] = useState<WeightEntry[]>([]);
  const [dt, setDT] = useState(() => localISODate(new Date()));
  const [value, setValue] = useState("");
  const [formErrors, setFormErrors] = useState<WeightFormErrors>({});
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [deletingDT, setDeletingDT] = useState("");
  const [entryToDelete, setEntryToDelete] = useState<WeightEntry | null>(null);
  const [loadError, setLoadError] = useState("");
  const [saveError, setSaveError] = useState("");

  const load = useCallback(async (period: Range) => {
    setLoading(true);
    setLoadError("");
    try {
      setItems(await getWeight(period.from, period.to));
    } catch (requestError) {
      setLoadError(requestError instanceof Error ? requestError.message : "Не удалось загрузить историю веса");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => { void load(defaultRange); }, [defaultRange, load]);

  const chartData = useMemo(() => [...items].reverse().map((entry) => ({
    ...entry,
    timestamp: parseLocalDate(entry.dt).getTime()
  })), [items]);

  function updateRange(field: keyof Range, nextValue: string) {
    setRange((current) => ({ ...current, [field]: nextValue }));
    setRangeErrors((current) => ({ ...current, [field]: undefined }));
  }

  async function handleFilter(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const errors: RangeErrors = {};
    if (!range.from) errors.from = "Укажите начальную дату";
    if (!range.to) errors.to = "Укажите конечную дату";
    if (range.from && range.to && range.from > range.to) errors.from = "Дата «от» должна быть не позже даты «до»";
    setRangeErrors(errors);
    if (Object.keys(errors).length > 0) return;
    setActiveRange(range);
    await load(range);
  }

  async function handleSave(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const numericValue = Number(value.replace(",", "."));
    const errors: WeightFormErrors = {};
    if (!dt) errors.dt = "Укажите дату";
    if (!value.trim()) errors.value = "Укажите вес";
    else if (!Number.isFinite(numericValue) || numericValue <= 0) errors.value = "Вес должен быть больше нуля";
    setFormErrors(errors);
    setSaveError("");
    if (Object.keys(errors).length > 0) return;

    setSaving(true);
    try {
      await saveWeight({ dt, value: numericValue });
      setValue("");
      await load(activeRange);
    } catch (requestError) {
      const requestErrors = requestFormErrors(requestError);
      if (Object.keys(requestErrors).length > 0) setFormErrors(requestErrors);
      else setSaveError(requestError instanceof Error ? requestError.message : "Не удалось сохранить вес");
    } finally {
      setSaving(false);
    }
  }

  async function handleDelete() {
    if (!entryToDelete) return;
    setDeletingDT(entryToDelete.dt);
    setLoadError("");
    try {
      await deleteWeight(entryToDelete.dt);
      await load(activeRange);
    } catch (requestError) {
      setLoadError(requestError instanceof Error ? requestError.message : "Не удалось удалить запись веса");
    } finally {
      setEntryToDelete(null);
      setDeletingDT("");
    }
  }

  return (
    <div className="page weight-page">
      <header className="page-header weight-page-header">
        <div>
          <p className="eyebrow">Динамика</p>
          <h1>Вес</h1>
          <p className="page-description">Следите за изменениями веса за выбранный период.</p>
        </div>
      </header>

      <section className="weight-panel period-panel" aria-labelledby="period-heading">
        <h2 id="period-heading">Период</h2>
        <form className="weight-filter-form" noValidate onSubmit={handleFilter}>
          <label className="field">От
            <input aria-invalid={Boolean(rangeErrors.from)} type="date" value={range.from} onChange={(event) => updateRange("from", event.target.value)} />
            <FieldError message={rangeErrors.from} />
          </label>
          <label className="field">До
            <input aria-invalid={Boolean(rangeErrors.to)} type="date" value={range.to} onChange={(event) => updateRange("to", event.target.value)} />
            <FieldError message={rangeErrors.to} />
          </label>
          <button className="button primary" disabled={loading} type="submit">Показать</button>
        </form>
      </section>

      <div className="weight-overview">
        <section className="weight-panel weight-entry-panel" aria-labelledby="add-weight-heading">
          <h2 id="add-weight-heading">Добавить измерение</h2>
          {saveError && <div className="error" role="alert">{saveError}</div>}
          <form className="weight-entry-form" noValidate onSubmit={handleSave}>
            <label className="field">Дата
              <input aria-invalid={Boolean(formErrors.dt)} type="date" value={dt} onChange={(event) => { setDT(event.target.value); setFormErrors((current) => ({ ...current, dt: undefined })); }} />
              <FieldError message={formErrors.dt} />
            </label>
            <label className="field">Вес, кг
              <input
                aria-invalid={Boolean(formErrors.value)}
                inputMode="decimal"
                pattern="[0-9]*[.,]?[0-9]*"
                placeholder="Например, 82,4"
                value={value}
                onChange={(event) => { setValue(normalizeNumericInput(event.target.value)); setFormErrors((current) => ({ ...current, value: undefined })); }}
              />
              <FieldError message={formErrors.value} />
            </label>
            <button className="button primary" disabled={saving} type="submit">{saving ? "Сохраняем…" : "Сохранить"}</button>
            <p className="form-hint">Новое значение заменит существующее измерение за эту дату.</p>
          </form>
        </section>

        <section className="weight-panel chart-panel" aria-labelledby="weight-chart-heading">
          <div className="section-heading">
            <h2 id="weight-chart-heading">Динамика веса</h2>
            <span className="weight-unit">кг</span>
          </div>
          {loading ? <p className="muted">Загрузка графика…</p> : loadError ? (
            <div className="error" role="alert">{loadError}</div>
          ) : chartData.length === 0 ? (
            <p className="chart-empty">За выбранный период измерений пока нет.</p>
          ) : (
            <div className="weight-chart" aria-label="График изменения веса">
              <ResponsiveContainer width="100%" height="100%" minWidth={0}>
                <LineChart data={chartData} accessibilityLayer margin={{ top: 12, right: 14, bottom: 2, left: 0 }}>
                  <CartesianGrid stroke="#e2ebe8" strokeDasharray="4 4" vertical={false} />
                  <XAxis
                    dataKey="timestamp"
                    domain={["dataMin", "dataMax"]}
                    minTickGap={28}
                    scale="time"
                    tickFormatter={(timestamp) => shortDateFormat.format(new Date(Number(timestamp)))}
                    type="number"
                  />
                  <YAxis domain={["dataMin - 1", "dataMax + 1"]} tickFormatter={(weight) => weightFormat.format(Number(weight))} width={46} />
                  <Tooltip
                    formatter={(weight) => [`${weightFormat.format(Number(weight))} кг`, "Вес"]}
                    labelFormatter={(timestamp) => dateFormat.format(new Date(Number(timestamp)))}
                  />
                  <Line activeDot={{ r: 6 }} dataKey="value" dot={{ r: 4 }} stroke="#167d74" strokeWidth={3} type="linear" />
                </LineChart>
              </ResponsiveContainer>
            </div>
          )}
        </section>
      </div>

      <section className="weight-history" aria-labelledby="weight-history-heading">
        <div className="section-heading">
          <h2 id="weight-history-heading">История</h2>
          <span className="food-count" aria-label={`Найдено измерений: ${items.length}`}>{items.length}</span>
        </div>
        {loading ? <p className="muted">Загрузка…</p> : loadError ? null : items.length === 0 ? (
          <p className="empty">Добавьте первое измерение за выбранный период.</p>
        ) : (
          <div className="weight-list">
            {items.map((entry) => (
              <article className="weight-row" key={entry.dt}>
                <time dateTime={entry.dt}>{formatDate(entry.dt)}</time>
                <strong>{weightFormat.format(entry.value)} <small>кг</small></strong>
                <button className="text-button danger" disabled={deletingDT === entry.dt} type="button" onClick={() => setEntryToDelete(entry)}>
                  {deletingDT === entry.dt ? "Удаляем…" : "Удалить"}
                </button>
              </article>
            ))}
          </div>
        )}
      </section>

      <ConfirmDialog
        busy={deletingDT !== ""}
        busyLabel="Удаляем…"
        confirmLabel="Удалить"
        dangerous
        onCancel={() => setEntryToDelete(null)}
        onConfirm={() => void handleDelete()}
        open={entryToDelete !== null}
        title="Удалить измерение?"
      >
        Запись за {entryToDelete ? formatDate(entryToDelete.dt) : "выбранную дату"} будет удалена без возможности восстановления.
      </ConfirmDialog>
    </div>
  );
}
