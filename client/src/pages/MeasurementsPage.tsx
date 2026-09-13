import { type FormEvent, useCallback, useEffect, useState } from "react";
import { createMeasurement, getMeasurements, type Measurement, type MeasurementType } from "../api";

const labels: Record<MeasurementType, string> = { weight: "Вес", pressure: "Давление", pulse: "Пульс" };
const units: Record<MeasurementType, string> = { weight: "кг", pressure: "мм рт. ст.", pulse: "уд/мин" };

export function MeasurementsPage() {
  const [measurements, setMeasurements] = useState<Measurement[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [type, setType] = useState<MeasurementType>("weight");
  const [value, setValue] = useState("");
  const [saving, setSaving] = useState(false);

  const load = useCallback(async () => {
    try {
      setError("");
      setMeasurements(await getMeasurements());
    } catch (requestError) {
      setError(requestError instanceof Error ? requestError.message : "Не удалось загрузить данные");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => { void load(); }, [load]);

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const numericValue = Number(value.replace(",", "."));
    if (!Number.isFinite(numericValue)) return;
    setSaving(true);
    setError("");
    try {
      const created = await createMeasurement({ type, value: numericValue, unit: units[type], measuredAt: new Date().toISOString() });
      setMeasurements((current) => [created, ...current]);
      setValue("");
    } catch (requestError) {
      setError(requestError instanceof Error ? requestError.message : "Не удалось сохранить данные");
    } finally {
      setSaving(false);
    }
  }

  return (
    <div className="page measurements-layout">
      <section>
        <p className="eyebrow">История здоровья</p>
        <h1>Измерения</h1>
        <p className="page-intro">Добавляйте показатели и следите за их изменением.</p>
        {error && <div className="error" role="alert">{error}</div>}
        {loading ? <p className="muted">Загрузка…</p> : measurements.length === 0 ? (
          <p className="empty">Измерений пока нет.</p>
        ) : (
          <div className="measurement-list">
            {measurements.map((item) => (
              <article className="measurement-row" key={item.id}>
                <div className={`type-icon ${item.type}`} aria-hidden="true">{labels[item.type][0]}</div>
                <div><strong>{labels[item.type]}</strong><span>{new Intl.DateTimeFormat("ru", { day: "numeric", month: "long", hour: "2-digit", minute: "2-digit" }).format(new Date(item.measuredAt))}</span></div>
                <p><b>{item.value.toLocaleString("ru")}</b> {item.unit}</p>
              </article>
            ))}
          </div>
        )}
      </section>
      <aside className="form-card">
        <h2>Новое измерение</h2>
        <form onSubmit={handleSubmit}>
          <label>Показатель
            <select value={type} onChange={(event) => setType(event.target.value as MeasurementType)}>
              {Object.entries(labels).map(([key, label]) => <option value={key} key={key}>{label}</option>)}
            </select>
          </label>
          <label>Значение
            <div className="value-field">
              <input inputMode="decimal" required value={value} onChange={(event) => setValue(event.target.value)} placeholder="Например, 72.4" aria-label="Значение" />
              <span>{units[type]}</span>
            </div>
          </label>
          <button className="button" disabled={saving} type="submit">{saving ? "Сохраняем…" : "Сохранить"}</button>
        </form>
      </aside>
    </div>
  );
}
