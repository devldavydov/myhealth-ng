import { type FormEvent, useCallback, useEffect, useMemo, useRef, useState } from "react";
import { Link, useSearchParams } from "react-router-dom";
import { deleteBundle, getBundles, type BundleSummary, type PageSize, type Pagination } from "../api";
import { catalogSearchParams, readCatalogParams, type CatalogParams } from "../catalogParams";
import { CatalogPagination } from "../components/CatalogPagination";
import { ConfirmDialog } from "../components/ConfirmDialog";

const numberFormat = new Intl.NumberFormat("ru", { maximumFractionDigits: 2 });
const initialPagination: Pagination = { page: 1, pageSize: 20, total: 0, totalPages: 0 };

export function BundlePage() {
  const [searchParams, setSearchParams] = useSearchParams();
  const catalog = useMemo(() => readCatalogParams(searchParams), [searchParams]);
  const [items, setItems] = useState<BundleSummary[]>([]);
  const [query, setQuery] = useState(catalog.query);
  const [pagination, setPagination] = useState<Pagination>(initialPagination);
  const [loading, setLoading] = useState(true);
  const [deletingKey, setDeletingKey] = useState("");
  const [bundleToDelete, setBundleToDelete] = useState<BundleSummary | null>(null);
  const [error, setError] = useState("");
  const requestSequence = useRef(0);

  const navigate = useCallback((params: CatalogParams, replace = false) => {
    setSearchParams(catalogSearchParams(params), { replace });
  }, [setSearchParams]);

  const load = useCallback(async (params: CatalogParams) => {
    const request = ++requestSequence.current;
    setLoading(true);
    setError("");
    try {
      const result = await getBundles(params);
      if (request !== requestSequence.current) return;
      setItems(result.items);
      setPagination(result.pagination);

      const lastPage = Math.max(1, result.pagination.totalPages);
      if (params.page > lastPage) navigate({ ...params, page: lastPage }, true);
    } catch (requestError) {
      if (request !== requestSequence.current) return;
      setError(requestError instanceof Error ? requestError.message : "Не удалось загрузить бандлы");
    } finally {
      if (request === requestSequence.current) setLoading(false);
    }
  }, [navigate]);

  useEffect(() => {
    const canonical = catalogSearchParams(catalog);
    if (canonical.toString() !== searchParams.toString()) setSearchParams(canonical, { replace: true });
  }, [catalog, searchParams, setSearchParams]);

  useEffect(() => { setQuery(catalog.query); }, [catalog.query]);
  useEffect(() => {
    void load(catalog);
    return () => { requestSequence.current += 1; };
  }, [catalog, load]);

  function handleSearch(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    navigate({ ...catalog, query: query.trim(), page: 1 });
  }

  async function handleDelete() {
    if (!bundleToDelete) return;
    const bundle = bundleToDelete;
    setDeletingKey(bundle.key);
    setError("");
    try {
      await deleteBundle(bundle.key);
      await load(catalog);
    } catch (requestError) {
      setError(requestError instanceof Error ? requestError.message : "Не удалось удалить бандл");
    } finally {
      setBundleToDelete(null);
      setDeletingKey("");
    }
  }

  function handlePageSizeChange(pageSize: PageSize) {
    navigate({ ...catalog, page: 1, pageSize });
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
          <span className="food-count" aria-label={`Найдено бандлов: ${pagination.total}`}>{pagination.total}</span>
        </div>
        <form className="search-form" role="search" onSubmit={handleSearch}>
          <input aria-label="Поиск бандлов" value={query} onChange={(event) => setQuery(event.target.value)} placeholder="Название" />
          <button className="button primary" type="submit">Найти</button>
          {catalog.query && (
            <button className="button secondary" type="button" onClick={() => navigate({ ...catalog, query: "", page: 1 })}>
              Сбросить
            </button>
          )}
        </form>

        {error && <div className="error" role="alert">{error}</div>}
        {loading ? <p className="muted">Загрузка…</p> : items.length === 0 ? (
          <p className="empty">{catalog.query ? "По вашему запросу ничего не найдено." : "Бандлов пока нет. Добавьте первый."}</p>
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

        {!loading && !error && (
          <CatalogPagination
            noun="бандлов"
            onPageChange={(page) => navigate({ ...catalog, page })}
            onPageSizeChange={handlePageSizeChange}
            pagination={pagination}
          />
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
