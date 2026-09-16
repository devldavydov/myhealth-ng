import { type FormEvent, useCallback, useEffect, useMemo, useRef, useState } from "react";
import { Link, useSearchParams } from "react-router-dom";
import {
  ApiError,
  deleteActiveCalories,
  getActiveCalories,
  clearJournalMeal,
  deleteJournalItem,
  getBundleByKey,
  getJournal,
  getSettings,
  saveJournal,
  saveActiveCalories,
  type ActiveCalories,
  type Food,
  type JournalDay,
  type JournalZone,
  type MealType
} from "../api";
import { ConfirmDialog } from "../components/ConfirmDialog";
import { FoodBundlePicker, type FoodBundleOption } from "../components/FoodBundlePicker";
import { TimedNotification } from "../components/TimedNotification";

type PendingFood = { meal: MealType; food: Food };
type DeleteTarget = { meal: MealType; food?: Food };

const numberFormat = new Intl.NumberFormat("ru", { maximumFractionDigits: 1 });
const dateFormat = new Intl.DateTimeFormat("ru", { day: "numeric", month: "long", year: "numeric" });

function localISODate(value: Date): string {
  const year = String(value.getFullYear()).padStart(4, "0");
  const month = String(value.getMonth() + 1).padStart(2, "0");
  const day = String(value.getDate()).padStart(2, "0");
  return `${year}-${month}-${day}`;
}

function parseLocalDate(value: string): Date {
  return new Date(`${value}T00:00:00`);
}

function validDate(value: string | null): value is string {
  if (!value || !/^\d{4}-\d{2}-\d{2}$/.test(value)) return false;
  const parsed = parseLocalDate(value);
  return !Number.isNaN(parsed.getTime()) && localISODate(parsed) === value;
}

function shiftDate(value: string, days: number): string {
  const result = parseLocalDate(value);
  result.setDate(result.getDate() + days);
  return localISODate(result);
}

function normalizeWeight(value: string): string {
  const normalized = value.replace(/[^0-9.,]/g, "");
  const separator = normalized.search(/[.,]/);
  if (separator === -1) return normalized;
  return normalized.slice(0, separator + 1) + normalized.slice(separator + 1).replace(/[.,]/g, "");
}

function activeCaloriesFieldError(error: unknown): string {
  if (!(error instanceof ApiError)) return "";
  return error.details.value?.[0] ?? "";
}

export function JournalPage() {
  const [searchParams, setSearchParams] = useSearchParams();
  const rawDate = searchParams.get("dt");
  const today = localISODate(new Date());
  const selectedDate = validDate(rawDate) ? rawDate : localISODate(new Date());
  const [day, setDay] = useState<JournalDay | null>(null);
  const [calorieLimit, setCalorieLimit] = useState<number | null>(null);
  const [activeCalories, setActiveCalories] = useState<ActiveCalories | null>(null);
  const [activeCaloriesValue, setActiveCaloriesValue] = useState("");
  const [activeCaloriesError, setActiveCaloriesError] = useState("");
  const [deleteActiveCaloriesOpen, setDeleteActiveCaloriesOpen] = useState(false);
  const [activeMeal, setActiveMeal] = useState<MealType | null>(null);
  const [pendingFood, setPendingFood] = useState<PendingFood | null>(null);
  const [weight, setWeight] = useState("");
  const [weightError, setWeightError] = useState("");
  const [message, setMessage] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(true);
  const [mutating, setMutating] = useState(false);
  const [deleteTarget, setDeleteTarget] = useState<DeleteTarget | null>(null);
  const dismissMessage = useCallback(() => setMessage(""), []);
  const loadRequestRef = useRef(0);

  useEffect(() => {
    if (rawDate !== selectedDate) setSearchParams({ dt: selectedDate }, { replace: true });
  }, [rawDate, selectedDate, setSearchParams]);

  const load = useCallback(async (dt: string) => {
    const requestID = ++loadRequestRef.current;
    setLoading(true);
    setError("");
    setActiveCaloriesError("");
    try {
      const [journal, settings, dailyActiveCalories] = await Promise.all([
        getJournal(dt),
        getSettings(),
        getActiveCalories(dt)
      ]);
      if (requestID !== loadRequestRef.current) return;
      setDay(journal);
      setCalorieLimit(settings.defaultDailyCalorieLimit);
      setActiveCalories(dailyActiveCalories);
      setActiveCaloriesValue(dailyActiveCalories ? String(dailyActiveCalories.value) : "");
    } catch (requestError) {
      if (requestID === loadRequestRef.current) {
        setError(requestError instanceof Error ? requestError.message : "Не удалось загрузить журнал");
      }
    } finally {
      if (requestID === loadRequestRef.current) setLoading(false);
    }
  }, []);

  useEffect(() => {
    setActiveMeal(null);
    setPendingFood(null);
    setMessage("");
    void load(selectedDate);
    setDeleteTarget(null);
    setDeleteActiveCaloriesOpen(false);
    return () => { loadRequestRef.current += 1; };
  }, [load, selectedDate]);

  const effectiveCalorieLimit = activeCalories?.value ?? calorieLimit;
  const limitPercent = useMemo(
    () => effectiveCalorieLimit ? (day?.totals.cal ?? 0) / effectiveCalorieLimit * 100 : null,
    [day, effectiveCalorieLimit]
  );

  function chooseDate(value: string) {
    if (validDate(value)) setSearchParams({ dt: value });
  }

  function zoneFor(meal: MealType): JournalZone | undefined {
    return day?.zones.find((zone) => zone.meal === meal);
  }

  async function saveDailyActiveCalories(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const value = Number(activeCaloriesValue.replace(",", "."));
    setActiveCaloriesError("");
    setError("");
    setMessage("");
    if (!activeCaloriesValue.trim()) {
      setActiveCaloriesError("Укажите активные калории");
      return;
    }
    if (!Number.isFinite(value) || value <= 0) {
      setActiveCaloriesError("Введите число больше нуля");
      return;
    }
    setMutating(true);
    try {
      const saved = await saveActiveCalories({ dt: selectedDate, value });
      setActiveCalories(saved);
      setActiveCaloriesValue(String(saved.value));
      setMessage("Активные калории за день сохранены.");
    } catch (requestError) {
      const fieldError = activeCaloriesFieldError(requestError);
      if (fieldError) setActiveCaloriesError(fieldError);
      else setError(requestError instanceof Error ? requestError.message : "Не удалось сохранить активные калории");
    } finally {
      setMutating(false);
    }
  }

  async function removeDailyActiveCalories() {
    setMutating(true);
    setError("");
    setMessage("");
    try {
      await deleteActiveCalories(selectedDate);
      setActiveCalories(null);
      setActiveCaloriesValue("");
      setMessage(calorieLimit ? "Дневное значение удалено. Используется лимит из настроек." : "Дневное значение удалено. Дневной лимит не задан.");
    } catch (requestError) {
      setError(requestError instanceof Error ? requestError.message : "Не удалось удалить активные калории");
    } finally {
      setDeleteActiveCaloriesOpen(false);
      setMutating(false);
    }
  }

  async function addOption(meal: MealType, option: FoodBundleOption) {
    setError("");
    setMessage("");
    setWeightError("");
    if (option.type === "food") {
      setPendingFood({ meal, food: option.food });
      setWeight("");
      return;
    }
    setMutating(true);
    try {
      const bundle = await getBundleByKey(option.key);
      const zone = zoneFor(meal);
      const replacements = bundle.items.filter((item) => zone?.items.some((current) => current.food.key === item.food.key)).length;
      const updated = await saveJournal({
        dt: selectedDate,
        meal,
        items: bundle.items.map((item) => ({ foodKey: item.food.key, weight: item.weight }))
      });
      setDay(updated);
      setMessage(`Бандл «${bundle.name}» добавлен${replacements ? `. Заменено позиций: ${replacements}.` : "."}`);
      setActiveMeal(null);
    } catch (requestError) {
      setError(requestError instanceof Error ? requestError.message : "Не удалось добавить бандл");
    } finally {
      setMutating(false);
    }
  }

  async function saveFood(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!pendingFood) return;
    const numericWeight = Number(weight.replace(",", "."));
    if (!weight.trim()) {
      setWeightError("Укажите вес");
      return;
    }
    if (!Number.isFinite(numericWeight) || numericWeight <= 0) {
      setWeightError("Вес должен быть больше нуля");
      return;
    }
    const replaced = Boolean(zoneFor(pendingFood.meal)?.items.some((item) => item.food.key === pendingFood.food.key));
    setMutating(true);
    setError("");
    try {
      const updated = await saveJournal({
        dt: selectedDate,
        meal: pendingFood.meal,
        items: [{ foodKey: pendingFood.food.key, weight: numericWeight }]
      });
      setDay(updated);
      setMessage(replaced ? `Вес продукта «${pendingFood.food.name}» заменён.` : `Продукт «${pendingFood.food.name}» добавлен.`);
      setPendingFood(null);
      setActiveMeal(null);
      setWeight("");
    } catch (requestError) {
      setError(requestError instanceof Error ? requestError.message : "Не удалось добавить продукт");
    } finally {
      setMutating(false);
    }
  }

  async function confirmDelete() {
    if (!deleteTarget) return;
    setMutating(true);
    setError("");
    try {
      const updated = deleteTarget.food
        ? await deleteJournalItem(selectedDate, deleteTarget.meal, deleteTarget.food.key)
        : await clearJournalMeal(selectedDate, deleteTarget.meal);
      setDay(updated);
      setMessage(deleteTarget.food ? `Продукт «${deleteTarget.food.name}» удалён.` : `Раздел «${deleteTarget.meal}» очищен.`);
    } catch (requestError) {
      setError(requestError instanceof Error ? requestError.message : "Не удалось изменить журнал");
    } finally {
      setDeleteTarget(null);
      setMutating(false);
    }
  }

  return (
    <div className="page journal-page">
      <header className="page-header journal-page-header">
        <div>
          <p className="eyebrow">Питание</p>
          <h1>Журнал</h1>
          <p className="page-description">Рацион за {dateFormat.format(parseLocalDate(selectedDate))}.</p>
        </div>
        <div className="journal-date-control">
          <button aria-label="Предыдущий день" className="button secondary" onClick={() => chooseDate(shiftDate(selectedDate, -1))} type="button">←</button>
          <input aria-label="Дата журнала" type="date" value={selectedDate} onChange={(event) => chooseDate(event.target.value)} />
          <button aria-label="Следующий день" className="button secondary" onClick={() => chooseDate(shiftDate(selectedDate, 1))} type="button">→</button>
          <button className="button secondary journal-today-button" disabled={selectedDate === today} onClick={() => chooseDate(today)} type="button">Сегодня</button>
        </div>
      </header>

      {error && <div className="error" role="alert">{error}</div>}
      <TimedNotification message={message} onDismiss={dismissMessage} />
      {loading || !day ? <p className="muted">Загрузка журнала…</p> : (
        <>
          <section className="journal-summary" aria-label="Итоги за день">
            <div className="journal-calories"><span>За день</span><strong>{numberFormat.format(day.totals.cal)} ккал</strong><small>{numberFormat.format(day.totals.weight)} г еды</small></div>
            <dl className="journal-macros">
              <div><dt>Белки</dt><dd>{numberFormat.format(day.totals.protein)} г</dd><small>{numberFormat.format(day.macroPercent.protein)}%</small></div>
              <div><dt>Жиры</dt><dd>{numberFormat.format(day.totals.fat)} г</dd><small>{numberFormat.format(day.macroPercent.fat)}%</small></div>
              <div><dt>Углеводы</dt><dd>{numberFormat.format(day.totals.carb)} г</dd><small>{numberFormat.format(day.macroPercent.carb)}%</small></div>
            </dl>
            <form className="active-calories-form" noValidate onSubmit={saveDailyActiveCalories}>
              <label>Активные калории за день
                <input
                  aria-label="Активные калории за день"
                  aria-invalid={Boolean(activeCaloriesError)}
                  inputMode="decimal"
                  pattern="[0-9]*[.,]?[0-9]*"
                  placeholder={calorieLimit ? `По умолчанию ${calorieLimit}` : "Например, 2500"}
                  value={activeCaloriesValue}
                  onChange={(event) => {
                    setActiveCaloriesValue(normalizeWeight(event.target.value));
                    setActiveCaloriesError("");
                    setMessage("");
                  }}
                />
                {activeCaloriesError && <span className="field-error" role="alert">{activeCaloriesError}</span>}
                <span className="form-hint">
                  {activeCalories ? "Используется значение за выбранный день." : calorieLimit ? `Сейчас применяется дефолт: ${numberFormat.format(calorieLimit)} ккал.` : "Дефолтный лимит тоже не задан."}
                </span>
              </label>
              <div className="active-calories-actions">
                <button className="button primary" disabled={mutating} type="submit">Сохранить</button>
                {activeCalories && <button className="button secondary" disabled={mutating} onClick={() => setDeleteActiveCaloriesOpen(true)} type="button">Убрать значение</button>}
              </div>
            </form>
            {effectiveCalorieLimit ? (
              <div aria-label="Прогресс дневного лимита" className={`calorie-goal ${(limitPercent ?? 0) > 100 ? "exceeded" : ""}`}>
                <div>
                  <span>{activeCalories ? "Лимит за выбранный день" : "Дефолтный дневной лимит"}</span>
                  <strong>{numberFormat.format(limitPercent ?? 0)}% · {numberFormat.format(day.totals.cal)} из {numberFormat.format(effectiveCalorieLimit)} ккал</strong>
                </div>
                <div className="calorie-goal-track"><span style={{ width: `${Math.min(100, limitPercent ?? 0)}%` }} /></div>
              </div>
            ) : <p className="calorie-goal-empty">Дневной лимит не задан. Задайте активные калории за день или <Link to="/settings">дефолтный лимит в настройках</Link>.</p>}
          </section>

          <div className="journal-zones">
            {day.zones.map((zone) => (
              <section className="journal-zone" key={zone.meal}>
                <header className="journal-zone-header">
                  <div><h2>{zone.meal}</h2><p>{numberFormat.format(zone.totals.cal)} ккал · Б {numberFormat.format(zone.totals.protein)} · Ж {numberFormat.format(zone.totals.fat)} · У {numberFormat.format(zone.totals.carb)}</p></div>
                  <div className="journal-zone-actions">
                    <button className="text-button" disabled={mutating} onClick={() => { setActiveMeal(activeMeal === zone.meal ? null : zone.meal); setPendingFood(null); setMessage(""); }} type="button">{activeMeal === zone.meal ? "Закрыть" : "Добавить"}</button>
                    {zone.items.length > 0 && <button className="text-button danger" disabled={mutating} onClick={() => setDeleteTarget({ meal: zone.meal })} type="button">Очистить</button>}
                  </div>
                </header>

                {activeMeal === zone.meal && (
                  <div className="journal-add-panel">
                    <FoodBundlePicker disabled={mutating} inputId={`journal-picker-${day.zones.indexOf(zone)}`} onError={setError} onSelect={(option) => addOption(zone.meal, option)} />
                    {pendingFood?.meal === zone.meal && (
                      <form className="journal-weight-form" noValidate onSubmit={saveFood}>
                        <div><strong>{pendingFood.food.name}</strong><small>{pendingFood.food.brand || "Без бренда"}</small></div>
                        <label>Вес, г
                          <input aria-invalid={Boolean(weightError)} aria-label={`Вес продукта ${pendingFood.food.name}`} inputMode="decimal" pattern="[0-9]*[.,]?[0-9]*" value={weight} onChange={(event) => { setWeight(normalizeWeight(event.target.value)); setWeightError(""); }} />
                          {weightError && <span className="field-error" role="alert">{weightError}</span>}
                        </label>
                        <button className="button primary" disabled={mutating} type="submit">Добавить</button>
                      </form>
                    )}
                  </div>
                )}

                {zone.items.length === 0 ? <p className="journal-zone-empty">Нет записей</p> : (
                  <div className="journal-items">
                    {zone.items.map((item) => (
                      <div className="journal-item" key={item.food.key}>
                        <div><strong>{item.food.name}</strong><small>{item.food.brand || "Без бренда"}</small></div>
                        <span>{numberFormat.format(item.weight)} г</span>
                        <span>{numberFormat.format(item.food.cal100 * item.weight / 100)} ккал</span>
                        <button aria-label={`Удалить ${item.food.name} из ${zone.meal}`} className="text-button danger" disabled={mutating} onClick={() => setDeleteTarget({ meal: zone.meal, food: item.food })} type="button">Удалить</button>
                      </div>
                    ))}
                  </div>
                )}
              </section>
            ))}
          </div>
        </>
      )}

      <ConfirmDialog
        busy={mutating}
        busyLabel="Удаляем…"
        confirmLabel="Убрать"
        dangerous
        onCancel={() => setDeleteActiveCaloriesOpen(false)}
        onConfirm={() => void removeDailyActiveCalories()}
        open={deleteActiveCaloriesOpen}
        title="Убрать активные калории за день?"
      >
        {calorieLimit ? "После удаления для этой даты будет использоваться дефолтный лимит из настроек." : "После удаления дневной лимит для этой даты не будет задан."}
      </ConfirmDialog>

      <ConfirmDialog
        busy={mutating}
        busyLabel="Удаляем…"
        confirmLabel={deleteTarget?.food ? "Удалить" : "Очистить"}
        dangerous
        onCancel={() => setDeleteTarget(null)}
        onConfirm={() => void confirmDelete()}
        open={deleteTarget !== null}
        title={deleteTarget?.food ? "Удалить продукт из журнала?" : "Очистить раздел?"}
      >
        {deleteTarget?.food
          ? `Продукт «${deleteTarget.food.name}» будет удалён из раздела «${deleteTarget.meal}».`
          : `Все продукты из раздела «${deleteTarget?.meal ?? ""}» будут удалены.`}
      </ConfirmDialog>
    </div>
  );
}
