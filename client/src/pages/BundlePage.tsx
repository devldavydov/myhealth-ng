import { type FormEvent, useCallback, useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { deleteBundle, getBundles, type BundleSummary } from "../api";
import { ConfirmDialog } from "../components/ConfirmDialog";

const numberFormat = new Intl.NumberFormat("ru", { maximumFractionDigits: 2 });

export function BundlePage() {
  const [items, setItems] = useState<BundleSummary[]>([]);
  const [query, setQuery] = useState("");
  const [activeQuery, setActiveQuery] = useState("");
  const [loading, setLoading] = useState(true);
  const [deletingKey, setDeletingKey] = useState("");
  const [bundleToDelete, setBundleToDelete] = useState<BundleSummary | null>(null);
  const [error, setError] = useState("");

  const load = useCallback(async (search: string) => {
    setLoading(true);
    setError("");
    try {
      setItems(await getBundles(search));
    } catch (requestError) {
      setError(requestError instanceof Error ? requestError.message : "Не удалось загрузить бандлы");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => { void load(""); }, [load]);

  async function handleSearch(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const normalized = query.trim();
    setActiveQuery(normalized);
    await load(normalized);
  }

  async function handleDelete() {
    if (!bundleToDelete) return;
    const bundle = bundleToDelete;
    setDeletingKey(bundle.key);
    setError("");
    try {
      await deleteBundle(bundle.key);
      await load(activeQuery);
    } catch (requestError) {
      setError(requestError instanceof Error ? requestError.message : "Не удалось удалить бандл");
    } finally {
      setBundleToDelete(null);
      setDeletingKey("");
    }
  }

  return (
    <div className="page bundle-page">
      <header className="page-header">
        <div>
          <p className="eyebrow">Наборы</p>
          <h1>Бандлы</h1>
          <p className="page-description">Готовые порции из нескольких продуктов.</p>
        </div>
        <Link className="button primary add-food-button" to="/bundle/new">Добавить бандл</Link>
      </header>

      <section aria-labelledby="bundle-list-heading">
        <div className="section-heading">
          <h2 id="bundle-list-heading">Список</h2>
          <span className="food-count" aria-label={`Найдено бандлов: ${items.length}`}>{items.length}</span>
        </div>
        <form className="search-form" role="search" onSubmit={handleSearch}>
          <input aria-label="Поиск бандлов" value={query} onChange={(event) => setQuery(event.target.value)} placeholder="Название" />
          <button className="button primary" type="submit">Найти</button>
          {activeQuery && (
            <button className="button secondary" type="button" onClick={() => { setQuery(""); setActiveQuery(""); void load(""); }}>
              Сбросить
            </button>
          )}
        </form>

        {error && <div className="error" role="alert">{error}</div>}
        {loading ? <p className="muted">Загрузка…</p> : items.length === 0 ? (
          <p className="empty">{activeQuery ? "По вашему запросу ничего не найдено." : "Бандлов пока нет. Добавьте первый."}</p>
        ) : (
          <div className="food-list">
            {items.map((bundle) => (
              <article className="food-card bundle-card" key={bundle.key}>
                <div className="food-card-heading">
                  <div>
                    <h3>{bundle.name}</h3>
                    <p>{bundle.itemCount} продуктов · {numberFormat.format(bundle.totals.weight)} г</p>
                  </div>
                  <strong>{numberFormat.format(bundle.totals.cal)} <small>ккал</small></strong>
                </div>
                <dl className="macros">
                  <div><dt>Белки</dt><dd>{numberFormat.format(bundle.totals.protein)} г</dd></div>
                  <div><dt>Жиры</dt><dd>{numberFormat.format(bundle.totals.fat)} г</dd></div>
                  <div><dt>Углеводы</dt><dd>{numberFormat.format(bundle.totals.carb)} г</dd></div>
                </dl>
                <div className="food-actions">
                  <Link className="text-button" to={`/bundle/${encodeURIComponent(bundle.key)}/edit`}>Редактировать</Link>
                  <button className="text-button danger" disabled={deletingKey === bundle.key} type="button" onClick={() => setBundleToDelete(bundle)}>
                    {deletingKey === bundle.key ? "Удаляем…" : "Удалить"}
                  </button>
                </div>
              </article>
            ))}
          </div>
        )}
      </section>

      <ConfirmDialog
        busy={deletingKey !== ""}
        busyLabel="Удаляем…"
        confirmLabel="Удалить"
        dangerous
        onCancel={() => setBundleToDelete(null)}
        onConfirm={() => void handleDelete()}
        open={bundleToDelete !== null}
        title="Удалить бандл?"
      >
        Бандл «{bundleToDelete?.name}» будет удалён. Продукты из него останутся в справочнике.
      </ConfirmDialog>
    </div>
  );
}
