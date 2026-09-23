import { type FormEvent, useCallback, useEffect, useState } from "react";
import { ApiError, getSettings, saveSettings } from "../api";
import { CalorieCalculatorDialog } from "../components/CalorieCalculatorDialog";
import { TimedNotification } from "../components/TimedNotification";

interface SettingsPageProps {
  userName: string;
}

function fieldErrorFrom(error: unknown): string {
  if (!(error instanceof ApiError)) return "";
  return error.details.defaultDailyCalorieLimit?.[0] ?? "";
}

export function SettingsPage({ userName }: SettingsPageProps) {
  const [limit, setLimit] = useState("");
  const [fieldError, setFieldError] = useState("");
  const [loadError, setLoadError] = useState("");
  const [saveError, setSaveError] = useState("");
  const [saved, setSaved] = useState(false);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [calculatorOpen, setCalculatorOpen] = useState(false);
  const dismissSaved = useCallback(() => setSaved(false), []);

  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    void getSettings()
      .then((settings) => {
        if (!cancelled) setLimit(settings.defaultDailyCalorieLimit === null ? "" : String(settings.defaultDailyCalorieLimit));
      })
      .catch((error: unknown) => {
        if (!cancelled) setLoadError(error instanceof Error ? error.message : "Не удалось загрузить настройки");
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });
    return () => { cancelled = true; };
  }, []);

  function changeLimit(value: string) {
    setLimit(value.replace(/\D/g, ""));
    setFieldError("");
    setSaveError("");
    setSaved(false);
  }

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const value = Number(limit);
    setSaveError("");
    setSaved(false);
    if (limit === "") {
      setFieldError("Укажите дневной лимит");
      return;
    }
    if (!Number.isInteger(value) || value < 1 || value > 10000) {
      setFieldError("Введите целое число от 1 до 10000");
      return;
    }

    setSaving(true);
    try {
      const settings = await saveSettings({ defaultDailyCalorieLimit: value });
      setLimit(settings.defaultDailyCalorieLimit === null ? "" : String(settings.defaultDailyCalorieLimit));
      setSaved(true);
    } catch (error) {
      const requestFieldError = fieldErrorFrom(error);
      if (requestFieldError) setFieldError(requestFieldError);
      else setSaveError(error instanceof Error ? error.message : "Не удалось сохранить настройки");
    } finally {
      setSaving(false);
    }
  }

  return (
    <div className="page settings-page">
      <header className="form-page-header">
        <p className="eyebrow">{userName}</p>
        <h1>Настройки</h1>
        <p className="page-description">Персональные параметры используются только для текущего пользователя.</p>
      </header>

      {loading ? <p className="muted">Загрузка настроек…</p> : loadError ? (
        <div className="error" role="alert">{loadError}</div>
      ) : (
        <section className="food-form-card settings-card">
          {saveError && <div className="error" role="alert">{saveError}</div>}
          <TimedNotification message={saved ? "Настройки сохранены." : ""} onDismiss={dismissSaved} />
          <form noValidate onSubmit={handleSubmit}>
            <div className="settings-limit-field">
              <label htmlFor="default-calorie-limit">Лимит ккал в день по умолчанию</label>
              <div className="settings-limit-control">
                <input
                  aria-invalid={Boolean(fieldError)}
                  id="default-calorie-limit"
                  inputMode="numeric"
                  pattern="[0-9]*"
                  value={limit}
                  onChange={(event) => changeLimit(event.target.value)}
                  placeholder="Например, 2000"
                />
                <button className="button secondary" onClick={() => setCalculatorOpen(true)} type="button">Рассчитать</button>
              </div>
              {fieldError && <span className="field-error" role="alert">{fieldError}</span>}
              <span className="form-hint">Целое число от 1 до 10000 ккал.</span>
            </div>
            <div className="form-actions">
              <button className="button primary" disabled={saving} type="submit">
                {saving ? "Сохраняем…" : "Сохранить"}
              </button>
            </div>
          </form>
        </section>
      )}
      <CalorieCalculatorDialog
        open={calculatorOpen}
        onCancel={() => setCalculatorOpen(false)}
        onSelect={(value) => {
          changeLimit(String(value));
          setCalculatorOpen(false);
        }}
      />
    </div>
  );
}
