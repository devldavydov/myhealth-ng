import { useCallback, useEffect, useRef } from "react";
import AsyncSelect from "react-select/async";
import { components, type GroupBase, type OptionProps } from "react-select";
import { getBundles, getFood, type BundleSummary, type Food } from "../api";

export type FoodBundleOption =
  | { type: "food"; key: string; label: string; food: Food }
  | { type: "bundle"; key: string; label: string; bundle: BundleSummary };

type SearchGroup = GroupBase<FoodBundleOption>;

interface FoodBundlePickerProps {
  disabled?: boolean;
  excludeBundleKey?: string;
  inputId?: string;
  onError: (message: string) => void;
  onSelect: (option: FoodBundleOption) => void | Promise<void>;
  placeholder?: string;
}

const numberFormat = new Intl.NumberFormat("ru", { maximumFractionDigits: 2 });

function SearchOptionView(props: OptionProps<FoodBundleOption, false, SearchGroup>) {
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

function groupLabel(label: string, shown: number, total: number): string {
  return total > shown ? `${label} · первые ${shown} из ${total}, уточните запрос` : label;
}

export function FoodBundlePicker({ disabled, excludeBundleKey, inputId = "food-bundle-picker", onError, onSelect, placeholder }: FoodBundlePickerProps) {
  const debounceTimer = useRef<number | undefined>(undefined);
  const pendingResolve = useRef<((groups: SearchGroup[]) => void) | null>(null);
  const requestSequence = useRef(0);

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
        const query = { query: inputValue, page: 1, pageSize: 50 as const };
        void Promise.all([getFood(query), getBundles(query)])
          .then(([foodPage, bundlePage]) => {
            if (sequence !== requestSequence.current) {
              resolve([]);
              return;
            }
            const bundles = bundlePage.items.filter((item) => item.key !== excludeBundleKey);
            const groups: SearchGroup[] = [
              {
                label: groupLabel("Еда", foodPage.items.length, foodPage.pagination.total),
                options: foodPage.items.map((item) => ({ type: "food" as const, key: item.key, label: item.name, food: item }))
              },
              {
                label: groupLabel("Бандлы", bundlePage.items.length, bundlePage.pagination.total),
                options: bundles.map((item) => ({ type: "bundle" as const, key: item.key, label: item.name, bundle: item }))
              }
            ];
            resolve(groups.filter((group) => group.options.length > 0));
          })
          .catch((error: unknown) => {
            if (sequence === requestSequence.current) onError(error instanceof Error ? error.message : "Не удалось выполнить поиск");
            resolve([]);
          });
      }, 250);
    });
  }, [excludeBundleKey, onError]);

  return (
    <AsyncSelect<FoodBundleOption, false, SearchGroup>
      aria-label="Поиск еды или бандла"
      cacheOptions
      classNamePrefix="bundle-select"
      components={{ Option: SearchOptionView }}
      defaultOptions
      getOptionValue={(option) => option.type + ":" + option.key}
      inputId={inputId}
      isClearable
      isDisabled={disabled}
      loadingMessage={() => "Ищем…"}
      loadOptions={loadOptions}
      noOptionsMessage={() => "Ничего не найдено"}
      onChange={(option) => { if (option) void onSelect(option); }}
      placeholder={placeholder ?? "Начните вводить название"}
      value={null}
    />
  );
}
