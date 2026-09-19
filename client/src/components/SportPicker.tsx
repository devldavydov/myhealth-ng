import { useCallback, useEffect, useRef } from "react";
import AsyncSelect from "react-select/async";
import { components, type OptionProps } from "react-select";
import { getSports, type Sport } from "../api";

export type SportOption = { value: string; label: string; sport: Sport };

type Props = {
  disabled?: boolean;
  onError: (message: string) => void;
  onSelect: (sport: Sport) => void;
  value?: Sport | null;
};

function OptionView(props: OptionProps<SportOption, false>) {
  return <components.Option {...props}><span className="search-option-copy"><strong>{props.data.sport.name}</strong><small>{props.data.sport.unit}{props.data.sport.comment ? ` · ${props.data.sport.comment}` : ""}</small></span></components.Option>;
}

export function SportPicker({ disabled, onError, onSelect, value }: Props) {
  const debounceTimer = useRef<number | undefined>(undefined);
  const sequence = useRef(0);
  useEffect(() => () => { if (debounceTimer.current !== undefined) clearTimeout(debounceTimer.current); sequence.current += 1; }, []);
  const loadOptions = useCallback((input: string) => new Promise<SportOption[]>((resolve) => {
    const request = ++sequence.current;
    if (debounceTimer.current !== undefined) clearTimeout(debounceTimer.current);
    debounceTimer.current = window.setTimeout(() => {
      void getSports({ query: input, page: 1, pageSize: 50 }).then((page) => {
        if (request !== sequence.current) return resolve([]);
        resolve(page.items.map((sport) => ({ value: sport.key, label: sport.name, sport })));
      }).catch((error: unknown) => { if (request === sequence.current) onError(error instanceof Error ? error.message : "Не удалось выполнить поиск"); resolve([]); });
    }, 250);
  }), [onError]);
  const selected = value ? { value: value.key, label: value.name, sport: value } : null;
  return <AsyncSelect<SportOption, false>
    aria-label="Поиск вида спорта"
    cacheOptions
    classNamePrefix="bundle-select"
    components={{ Option: OptionView }}
    defaultOptions
    isClearable
    isDisabled={disabled}
    loadingMessage={() => "Ищем…"}
    loadOptions={loadOptions}
    menuPlacement="bottom"
    menuPortalTarget={document.body}
    menuPosition="absolute"
    noOptionsMessage={() => "Ничего не найдено"}
    onChange={(option) => { if (option) onSelect(option.sport); }}
    placeholder="Начните вводить название"
    styles={{ menuPortal: (base) => ({ ...base, zIndex: 1000 }) }}
    value={selected}
  />;
}
