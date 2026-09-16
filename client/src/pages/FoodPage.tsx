import { type FormEvent, useCallback, useEffect, useMemo, useRef, useState } from "react";
import { Link, useSearchParams } from "react-router-dom";
import { deleteFood, getFood, type Food, type PageSize, type Pagination } from "../api";
import { catalogSearchParams, readCatalogParams, type CatalogParams } from "../catalogParams";
import { CatalogPagination } from "../components/CatalogPagination";
import { ConfirmDialog } from "../components/ConfirmDialog";

const numberFormat = new Intl.NumberFormat("ru", { maximumFractionDigits: 2 });
const initialPagination: Pagination = { page: 1, pageSize: 20, total: 0, totalPages: 0 };

export function FoodPage() {
  const [searchParams, setSearchParams] = useSearchParams();
  const catalog = useMemo(() => readCatalogParams(searchParams), [searchParams]);
  const [items, setItems] = useState<Food[]>([]);
  const [query, setQuery] = useState(catalog.query);
  const [pagination, setPagination] = useState<Pagination>(initialPagination);
  const [loading, setLoading] = useState(true);
  const [deletingKey, setDeletingKey] = useState("");
  const [foodToDelete, setFoodToDelete] = useState<Food | null>(null);
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
      const result = await getFood(params);
      if (request !== requestSequence.current) return;
      setItems(result.items);
      setPagination(result.pagination);

      const lastPage = Math.max(1, result.pagination.totalPages);
      if (params.page > lastPage) navigate({ ...params, page: lastPage }, true);
    } catch (requestError) {
      if (request !== requestSequence.current) return;
      setError(requestError instanceof Error ? requestError.message : "Не удалось загрузить продукты");
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
    if (!foodToDelete) return;
    const food = foodToDelete;
    setDeletingKey(food.key);
    setError("");
    try {
      await deleteFood(food.key);
      await load(catalog);
    } catch (requestError) {
      setError(requestError instanceof Error ? requestError.message : "Не удалось удалить продукт");
    } finally {
      setFoodToDelete(null);
      setDeletingKey("");
    }
  }

  function handlePageSizeChange(pageSize: PageSize) {
    navigate({ ...catalog, page: 1, pageSize });
  }

  return (
    <div className="page food-page">
      <header className="page-header">
        <div>
          <p className="eyebrow">Справочник</p>
          <h1>Продукты</h1>
          <p className="page-description">Калорийность и КБЖУ продуктов на 100 грамм.</p>
        </div>
        <Link className="button primary add-food-button" to="/food/new">Добавить продукт</Link>
      </header>

      <section className="food-catalog" aria-labelledby="food-list-heading">
        <div className="section-heading">
          <h2 id="food-list-heading">Список</h2>
          <span className="food-count" aria-label={`Найдено продуктов: ${pagination.total}`}>{pagination.total}</span>
        </div>
        <form className="search-form" role="search" onSubmit={handleSearch}>
          <input aria-label="Поиск продуктов" value={query} onChange={(event) => setQuery(event.target.value)} placeholder="Название или бренд" />
          <button className="button primary" type="submit">Найти</button>
          {catalog.query && (
            <button className="button secondary" type="button" onClick={() => navigate({ ...catalog, query: "", page: 1 })}>
              Сбросить
            </button>
          )}
        </form>

        {error && <div className="error" role="alert">{error}</div>}
        {loading ? <p className="muted">Загрузка…</p> : items.length === 0 ? (
          <p className="empty">{catalog.query ? "По вашему запросу ничего не найдено." : "Продуктов пока нет. Добавьте первый."}</p>
        ) : (
          <div className="food-list">
            {items.map((food) => (
              <article className="food-card" key={food.key}>
                <div className="food-card-heading">
                  <div><h3>{food.name}</h3><p>{food.brand || "Без бренда"}</p></div>
                  <strong>{numberFormat.format(food.cal100)} <small>ккал</small></strong>
                </div>
                <dl className="macros">
                  <div><dt>Белки</dt><dd>{numberFormat.format(food.prot100)} г</dd></div>
                  <div><dt>Жиры</dt><dd>{numberFormat.format(food.fat100)} г</dd></div>
                  <div><dt>Углеводы</dt><dd>{numberFormat.format(food.carb100)} г</dd></div>
                </dl>
                {food.comment && <p className="food-comment">{food.comment}</p>}
                <div className="food-actions">
                  <Link className="text-button" to={`/food/${encodeURIComponent(food.key)}/edit`}>Редактировать</Link>
                  <button className="text-button danger" disabled={deletingKey === food.key} type="button" onClick={() => setFoodToDelete(food)}>
                    {deletingKey === food.key ? "Удаляем…" : "Удалить"}
                  </button>
                </div>
              </article>
            ))}
          </div>
        )}

        {!loading && !error && (
          <CatalogPagination
            noun="продуктов"
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
        onCancel={() => setFoodToDelete(null)}
        onConfirm={() => void handleDelete()}
        open={foodToDelete !== null}
        title="Удалить продукт?"
      >
        Продукт «{foodToDelete?.name}» будет удалён без возможности восстановления.
      </ConfirmDialog>
    </div>
  );
}
