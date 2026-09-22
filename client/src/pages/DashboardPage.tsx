import { type FormEvent, useCallback, useEffect, useMemo, useState } from "react";
import { ArrowDown, ArrowUp, Minus } from "lucide-react";
import { Link } from "react-router-dom";
import {
  Bar,
  BarChart,
  CartesianGrid,
  ReferenceLine,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis
} from "recharts";
import { getDashboard, type DashboardData } from "../api";
import { categoricalYearMarkers, ChartYearLines } from "../components/ChartYearLines";

type Range = { from: string; to: string };
type RangeErrors = Partial<Record<keyof Range, string>>;

const calorieFormat = new Intl.NumberFormat("ru", { maximumFractionDigits: 1 });
const weightFormat = new Intl.NumberFormat("ru", { maximumFractionDigits: 2 });
const activityFormat = new Intl.NumberFormat("ru", { maximumFractionDigits: 2 });
const fullDateFormat = new Intl.DateTimeFormat("ru", { day: "numeric", month: "long", year: "numeric" });
const shortDateFormat = new Intl.DateTimeFormat("ru", { day: "2-digit", month: "2-digit" });

function localISODate(value: Date): string {
  const year = String(value.getFullYear()).padStart(4, "0");
  const month = String(value.getMonth() + 1).padStart(2, "0");
  const day = String(value.getDate()).padStart(2, "0");
  return `${year}-${month}-${day}`;
}

function parseLocalDate(value: string): Date {
  return new Date(`${value}T00:00:00`);
}

export function initialDashboardRange(today = new Date()): Range {
  const targetMonth = new Date(today.getFullYear(), today.getMonth() - 1, 1);
  const lastTargetDay = new Date(targetMonth.getFullYear(), targetMonth.getMonth() + 1, 0).getDate();
  const from = new Date(targetMonth.getFullYear(), targetMonth.getMonth(), Math.min(today.getDate(), lastTargetDay));
  return { from: localISODate(from), to: localISODate(today) };
}

function FieldError({ message }: { message?: string }) {
  return message ? <span className="field-error" role="alert">{message}</span> : null;
}

export function calorieBarFill(balance: number): string {
  return balance >= 0 ? "#2b9185" : "#c9535e";
}

function averageLabel(average: number): string {
  if (average > 0) return `В среднем дефицит ${calorieFormat.format(average)} ккал в день`;
  if (average < 0) return `В среднем перерасход ${calorieFormat.format(Math.abs(average))} ккал в день`;
  return "В среднем баланс калорий равен нулю";
}

function WeightChange({ value }: { value: number | null }) {
  if (value === null) {
    return <p className="dashboard-empty">Недостаточно измерений веса за выбранный период.</p>;
  }
  if (value > 0) {
    return <div className="weight-change increase"><ArrowUp aria-hidden="true" size={34} /><strong>+{weightFormat.format(value)} кг</strong><span>Вес увеличился</span></div>;
  }
  if (value < 0) {
    return <div className="weight-change decrease"><ArrowDown aria-hidden="true" size={34} /><strong>−{weightFormat.format(Math.abs(value))} кг</strong><span>Вес уменьшился</span></div>;
  }
  return <div className="weight-change neutral"><Minus aria-hidden="true" size={34} /><strong>0 кг</strong><span>Вес не изменился</span></div>;
}

export function DashboardPage() {
  const [defaultRange] = useState(initialDashboardRange);
  const [range, setRange] = useState(defaultRange);
  const [rangeErrors, setRangeErrors] = useState<RangeErrors>({});
  const [data, setData] = useState<DashboardData | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  const load = useCallback(async (period: Range) => {
    setLoading(true);
    setError("");
    try {
      setData(await getDashboard(period.from, period.to));
    } catch (reason) {
      setData(null);
      setError(reason instanceof Error ? reason.message : "Не удалось загрузить дашборд");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => { void load(defaultRange); }, [defaultRange, load]);

  const chartData = useMemo(() => (data?.calories?.days ?? []).map((day) => ({
    ...day,
    fill: calorieBarFill(day.balance)
  })), [data]);
  const yearMarkers = useMemo(() => categoricalYearMarkers(chartData), [chartData]);

  function updateRange(field: keyof Range, value: string) {
    setRange((current) => ({ ...current, [field]: value }));
    setRangeErrors((current) => ({ ...current, [field]: undefined }));
  }

  function handleFilter(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const errors: RangeErrors = {};
    if (!range.from) errors.from = "Укажите начальную дату";
    if (!range.to) errors.to = "Укажите конечную дату";
    if (range.from && range.to && range.from > range.to) errors.from = "Дата «от» должна быть не позже даты «до»";
    setRangeErrors(errors);
    if (Object.keys(errors).length === 0) void load(range);
  }

  return (
    <div className="page dashboard-page">
      <header className="page-header dashboard-header">
        <div>
          <p className="eyebrow">Обзор</p>
          <h1>Аналитика</h1>
          <p className="page-description">Ваши ключевые показатели здоровья за выбранный период.</p>
        </div>
      </header>

      <section className="weight-panel period-panel" aria-labelledby="dashboard-period-heading">
        <h2 id="dashboard-period-heading">Период</h2>
        <form className="weight-filter-form" noValidate onSubmit={handleFilter}>
          <label className="field date-field">От
            <input aria-invalid={Boolean(rangeErrors.from)} type="date" value={range.from} onChange={(event) => updateRange("from", event.target.value)} />
            <FieldError message={rangeErrors.from} />
          </label>
          <label className="field date-field">До
            <input aria-invalid={Boolean(rangeErrors.to)} type="date" value={range.to} onChange={(event) => updateRange("to", event.target.value)} />
            <FieldError message={rangeErrors.to} />
          </label>
          <button className="button primary" disabled={loading} type="submit">Показать</button>
        </form>
      </section>

      {error && <div className="error" role="alert">{error}</div>}
      {loading ? <p className="muted">Загрузка показателей…</p> : data && (
        <div className="dashboard-grid">
          <section className="weight-panel dashboard-calories" aria-labelledby="calorie-chart-heading">
            <div className="section-heading">
              <h2 id="calorie-chart-heading">Статистика по ккал</h2>
              <span className="weight-unit">ккал</span>
            </div>
            {data.calories === null ? (
              <p className="dashboard-callout">Чтобы построить график, задайте <Link to="/settings">дефолтный лимит ккал в настройках</Link>.</p>
            ) : chartData.length === 0 ? (
              <p className="chart-empty">За выбранный период в журнале питания нет записей.</p>
            ) : (
              <>
                <div className="dashboard-chart" aria-label="График баланса калорий">
                  <ResponsiveContainer width="100%" height="100%" minWidth={0}>
                    <BarChart data={chartData} accessibilityLayer margin={{ top: 24, right: 12, bottom: 2, left: 0 }}>
                      <CartesianGrid stroke="#e2ebe8" strokeDasharray="4 4" vertical={false} />
                      <XAxis dataKey="dt" minTickGap={28} tickFormatter={(value) => shortDateFormat.format(parseLocalDate(String(value)))} />
                      <YAxis tickFormatter={(value) => calorieFormat.format(Number(value))} width={54} />
                      <Tooltip
                        formatter={(value) => [`${calorieFormat.format(Number(value))} ккал`, "Баланс"]}
                        labelFormatter={(value) => fullDateFormat.format(parseLocalDate(String(value)))}
                      />
                      <ReferenceLine stroke="#81948f" y={0} />
                      <ChartYearLines categorical markers={yearMarkers} />
                      <Bar dataKey="balance" fill="#2b9185" maxBarSize={42} radius={[5, 5, 5, 5]} />
                    </BarChart>
                  </ResponsiveContainer>
                </div>
                {data.calories.average !== null && <p className={`dashboard-average ${data.calories.average < 0 ? "over" : "deficit"}`}>{averageLabel(data.calories.average)}</p>}
              </>
            )}
          </section>

          <section className="weight-panel dashboard-weight" aria-labelledby="weight-change-heading">
            <h2 id="weight-change-heading">Динамика веса</h2>
            <WeightChange value={data.weightChange} />
          </section>

          <section className="weight-panel dashboard-activities" aria-labelledby="top-activities-heading">
            <div className="section-heading">
              <h2 id="top-activities-heading">Топ-10 активности</h2>
              <span className="food-count" aria-label={`Видов активности: ${data.activities.length}`}>{data.activities.length}</span>
            </div>
            {data.activities.length === 0 ? <p className="dashboard-empty">За выбранный период спортивной активности пока нет.</p> : (
              <ol className="dashboard-activity-list">
                {data.activities.map((item) => (
                  <li key={item.sportKey}>
                    <strong>{item.name}</strong>
                    <span>{item.count} раз</span>
                    <b>{activityFormat.format(item.total)} {item.unit}</b>
                  </li>
                ))}
              </ol>
            )}
          </section>
        </div>
      )}
    </div>
  );
}
