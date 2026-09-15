import { type FormEvent, useCallback, useEffect, useMemo, useRef, useState } from "react";
import { Link, useNavigate, useParams } from "react-router-dom";
import AsyncSelect from "react-select/async";
import { components, type GroupBase, type OptionProps } from "react-select";
import {
  ApiError,
  createBundle,
  getBundleByKey,
  getBundles,
  getFood,
  updateBundle,
  type BundleItem,
  type BundleSummary,
  type Food
} from "../api";

type FormItem = { food: Food; weight: string };
type SearchOption =
  | { type: "food"; key: string; label: string; food: Food }
  | { type: "bundle"; key: string; label: string; bundle: BundleSummary };
type SearchGroup = GroupBase<SearchOption>;

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

export function mergeBundleItems(current: FormItem[], added: BundleItem[]): FormItem[] {
  const result = current.map((item) => ({ ...item }));
  const indexes = new Map(result.map((item, index) => [item.food.key, index]));
  for (const item of added) {
    const existingIndex = indexes.get(item.food.key);
    if (existingIndex === undefined) {
      indexes.set(item.food.key, result.length);
      result.push({ food: item.food, weight: String(item.weight) });
      continue;
    }
    const existing = parseNumber(result[existingIndex].weight);
    result[existingIndex] = {
      ...result[existingIndex],
      weight: String((Number.isFinite(existing) ? existing : 0) + item.weight)
    };
  }
  return result;
}

function SearchOptionView(props: OptionProps<SearchOption, false, SearchGroup>) {
  const option = props.data;
  const detail = option.type === "food"
    ? option.food.brand || "Без бренда"
    : option.bundle.itemCount + " продуктов · " + numberFormat.format(option.bundle.totals.weight) + " г";
  return (
    <components.Option {...props}>
      <span className={"search-option-kind " + option.type}>{option.type === "food" ? "Еда" : "Бандл"}</span>
      <span className="search-option-copy"><strong>{option.label}</strong><small>{detail}</small></span>
    </components.Option>
  );
}

export function BundleFormPage() {
  const { key } = useParams<{ key: string }>();
  const navigate = useNavigate();
  const editing = key !== undefined;
  const [name, setName] = useState("");
  const [items, setItems] = useState<FormItem[]>([]);
  const [nameError, setNameError] = useState("");
  const [itemsError, setItemsError] = useState("");
  const [weightErrors, setWeightErrors] = useState<Record<string, string>>({});
  const [loading, setLoading] = useState(editing);
  const [saving, setSaving] = useState(false);
  const [adding, setAdding] = useState(false);
  const [loadError, setLoadError] = useState("");
  const [error, setError] = useState("");
  const [pickerMessage, setPickerMessage] = useState("");
  const debounceTimer = useRef<number | undefined>(undefined);
  const pendingResolve = useRef<((groups: SearchGroup[]) => void) | null>(null);
  const requestSequence = useRef(0);

  useEffect(() => {
    if (!key) return;
    let cancelled = false;
    setLoading(true);
    void getBundleByKey(key)
      .then((bundle) => {
        if (cancelled) return;
        setName(bundle.name);
        setItems(bundle.items.map((item) => ({ food: item.food, weight: String(item.weight) })));
      })
      .catch((requestError: unknown) => {
        if (!cancelled) setLoadError(requestError instanceof Error ? requestError.message : "Не удалось загрузить бандл");
      })
      .finally(() => { if (!cancelled) setLoading(false); });
    return () => { cancelled = true; };
  }, [key]);

  useEffect(() => () => {
    if (debounceTimer.current !== undefined) window.clearTimeout(debounceTimer.current);
    pendingResolve.current?.([]);
    requestSequence.current++;
  }, []);

  const loadOptions = useCallback((inputValue: string): Promise<SearchGroup[]> => {
    const sequence = ++requestSequence.current;
    if (debounceTimer.current !== undefined) {
      window.clearTimeout(debounceTimer.current);
      pendingResolve.current?.([]);
    }
    return new Promise((resolve) => {
      pendingResolve.current = resolve;
      debounceTimer.current = window.setTimeout(() => {
        debounceTimer.current = undefined;
        pendingResolve.current = null;
        void Promise.all([getFood(inputValue), getBundles(inputValue)])
          .then(([food, bundles]) => {
            if (sequence !== requestSequence.current) {
              resolve([]);
              return;
            }
            const groups: SearchGroup[] = [
              {
                label: "Еда",
                options: food.map((item) => ({ type: "food" as const, key: item.key, label: item.name, food: item }))
              },
              {
                label: "Бандлы",
                options: bundles
                  .filter((item) => item.key !== key)
                  .map((item) => ({ type: "bundle" as const, key: item.key, label: item.name, bundle: item }))
              }
            ];
            resolve(groups.filter((group) => group.options.length > 0));
          })
          .catch((requestError: unknown) => {
            if (sequence === requestSequence.current) {
              setError(requestError instanceof Error ? requestError.message : "Не удалось выполнить поиск");
            }
            resolve([]);
          });
      }, 250);
    });
  }, [key]);

  async function addOption(option: SearchOption | null) {
    if (!option) return;
    setPickerMessage("");
    setItemsError("");
    if (option.type === "food") {
      if (items.some((item) => item.food.key === option.key)) {
        setPickerMessage("Этот продукт уже есть в составе.");
        return;
      }
      setItems((current) => [...current, { food: option.food, weight: "" }]);
      return;
    }
    setAdding(true);
    setError("");
    try {
      const nested = await getBundleByKey(option.key);
      setItems((current) => mergeBundleItems(current, nested.items));
      setPickerMessage("Бандл «" + nested.name + "» развёрнут в продукты.");
    } catch (requestError) {
      setError(requestError instanceof Error ? requestError.message : "Не удалось добавить бандл");
    } finally {
      setAdding(false);
    }
  }

  const totals = useMemo(() => items.reduce((result, item) => {
    const weight = parseNumber(item.weight);
    if (!Number.isFinite(weight) || weight <= 0) return result;
    const multiplier = weight / 100;
    result.weight += weight;
    result.cal += item.food.cal100 * multiplier;
    result.protein += item.food.prot100 * multiplier;
    result.fat += item.food.fat100 * multiplier;
    result.carb += item.food.carb100 * multiplier;
    return result;
  }, { weight: 0, cal: 0, protein: 0, fat: 0, carb: 0 }), [items]);

  function setWeight(foodKey: string, value: string) {
    setItems((current) => current.map((item) => item.food.key === foodKey ? { ...item, weight: normalizeNumericInput(value) } : item));
    setWeightErrors((current) => {
      if (!current[foodKey]) return current;
      const next = { ...current };
      delete next[foodKey];
      return next;
    });
  }

  function removeItem(foodKey: string) {
    setItems((current) => current.filter((item) => item.food.key !== foodKey));
    setWeightErrors((current) => {
      const next = { ...current };
      delete next[foodKey];
      return next;
    });
  }

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError("");
    let invalid = false;
    if (!name.trim()) {
      setNameError("Введите название бандла");
      invalid = true;
    }
    if (items.length === 0) {
      setItemsError("Добавьте хотя бы один продукт");
      invalid = true;
    }
    const nextWeightErrors: Record<string, string> = {};
    for (const item of items) {
      const weight = parseNumber(item.weight);
      if (item.weight.trim() === "") nextWeightErrors[item.food.key] = "Укажите вес";
      else if (!Number.isFinite(weight) || weight <= 0) nextWeightErrors[item.food.key] = "Вес должен быть больше нуля";
    }
    setWeightErrors(nextWeightErrors);
    if (Object.keys(nextWeightErrors).length > 0 || invalid) return;

    const data = {
      name: name.trim(),
      items: items.map((item) => ({ foodKey: item.food.key, weight: parseNumber(item.weight) }))
    };
    setSaving(true);
    try {
      if (key) await updateBundle(key, data);
      else await createBundle(data);
      navigate("/bundle");
    } catch (requestError) {
      if (requestError instanceof ApiError) {
        if (requestError.details.name?.[0]) setNameError(requestError.details.name[0]);
        if (requestError.details.items?.[0]) setItemsError(requestError.details.items[0]);
        if (requestError.details.name || requestError.details.items) return;
      }
      setError(requestError instanceof Error ? requestError.message : "Не удалось сохранить бандл");
    } finally {
      setSaving(false);
    }
  }

  return (
    <div className="page food-form-page bundle-form-page">
      <Link className="back-link" to="/bundle">← К бандлам</Link>
      <header className="form-page-header">
        <p className="eyebrow">{editing ? "Редактирование" : "Новый набор"}</p>
        <h1>{editing ? "Изменить бандл" : "Добавить бандл"}</h1>
      </header>

      {loading ? <p className="muted">Загрузка бандла…</p> : loadError ? (
        <div className="error" role="alert">{loadError}</div>
      ) : (
        <section className="food-form-card">
          {error && <div className="error" role="alert">{error}</div>}
          <form noValidate onSubmit={handleSubmit}>
            <label>Название
              <input aria-invalid={Boolean(nameError)} value={name} onChange={(event) => { setName(event.target.value); setNameError(""); }} />
              {nameError && <span className="field-error" role="alert">{nameError}</span>}
            </label>

            <div className="bundle-picker-field">
              <label htmlFor="bundle-item-picker">Добавить продукт или бандл</label>
              <AsyncSelect<SearchOption, false, SearchGroup>
                aria-label="Поиск еды или бандла"
                cacheOptions
                classNamePrefix="bundle-select"
                components={{ Option: SearchOptionView }}
                defaultOptions
                getOptionValue={(option) => option.type + ":" + option.key}
                inputId="bundle-item-picker"
                isClearable
                isDisabled={saving || adding}
                loadingMessage={() => "Ищем…"}
                loadOptions={loadOptions}
                noOptionsMessage={() => "Ничего не найдено"}
                onChange={(option) => void addOption(option)}
                placeholder={adding ? "Добавляем бандл…" : "Начните вводить название"}
                value={null}
              />
              {pickerMessage && <p className="form-hint" aria-live="polite">{pickerMessage}</p>}
              {itemsError && <span className="field-error" role="alert">{itemsError}</span>}
            </div>

            {items.length > 0 && (
              <div className="bundle-items">
                {items.map((item) => (
                  <div className="bundle-item-row" key={item.food.key}>
                    <div className="bundle-item-copy">
                      <strong>{item.food.name}</strong>
                      <small>{item.food.brand || "Без бренда"} · {numberFormat.format(item.food.cal100)} ккал/100 г</small>
                    </div>
                    <label>Вес, г
                      <input
                        aria-invalid={Boolean(weightErrors[item.food.key])}
                        aria-label={"Вес продукта " + item.food.name}
                        inputMode="decimal"
                        pattern="[0-9]*[.,]?[0-9]*"
                        value={item.weight}
                        onChange={(event) => setWeight(item.food.key, event.target.value)}
                      />
                      {weightErrors[item.food.key] && <span className="field-error" role="alert">{weightErrors[item.food.key]}</span>}
                    </label>
                    <button className="text-button danger" type="button" onClick={() => removeItem(item.food.key)}>Убрать</button>
                  </div>
                ))}
              </div>
            )}

            <div className="bundle-total" aria-live="polite">
              <strong>{numberFormat.format(totals.weight)} г · {numberFormat.format(totals.cal)} ккал</strong>
              <span>Б {numberFormat.format(totals.protein)} · Ж {numberFormat.format(totals.fat)} · У {numberFormat.format(totals.carb)}</span>
            </div>

            <div className="form-actions">
              <button className="button primary" disabled={saving || adding} type="submit">{saving ? "Сохраняем…" : editing ? "Сохранить изменения" : "Добавить"}</button>
              <Link className="button secondary" to="/bundle">Отмена</Link>
            </div>
          </form>
        </section>
      )}
    </div>
  );
}
