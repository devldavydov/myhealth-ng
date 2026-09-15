import { type FormEvent, useCallback, useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { deleteFood, getFood, type Food } from "../api";
import { ConfirmDialog } from "../components/ConfirmDialog";

const numberFormat = new Intl.NumberFormat("ru", { maximumFractionDigits: 2 });

export function FoodPage() {
  const [items, setItems] = useState<Food[]>([]);
  const [query, setQuery] = useState("");
  const [activeQuery, setActiveQuery] = useState("");
  const [loading, setLoading] = useState(true);
  const [deletingKey, setDeletingKey] = useState("");
  const [foodToDelete, setFoodToDelete] = useState<Food | null>(null);
  const [error, setError] = useState("");

  const load = useCallback(async (search: string) => {
    setLoading(true);
    setError("");
    try {
      setItems(await getFood(search));
    } catch (requestError) {
      setError(requestError instanceof Error ? requestError.message : "Не удалось загрузить продукты");
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
    if (!foodToDelete) return;
    const food = foodToDelete;
    setDeletingKey(food.key);
    setError("");
    try {
      await deleteFood(food.key);
      await load(activeQuery);
    } catch (requestError) {
      setError(requestError instanceof Error ? requestError.message : "Не удалось удалить продукт");
    } finally {
      setFoodToDelete(null);
      setDeletingKey("");
    }
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
          <span className="food-count" aria-label={`Найдено продуктов: ${items.length}`}>{items.length}</span>
        </div>
        <form className="search-form" role="search" onSubmit={handleSearch}>
          <input aria-label="Поиск продуктов" value={query} onChange={(event) => setQuery(event.target.value)} placeholder="Название или бренд" />
          <button className="button primary" type="submit">Найти</button>
          {activeQuery && (
            <button className="button secondary" type="button" onClick={() => { setQuery(""); setActiveQuery(""); void load(""); }}>
              Сбросить
            </button>
          )}
        </form>

        {error && <div className="error" role="alert">{error}</div>}
        {loading ? <p className="muted">Загрузка…</p> : items.length === 0 ? (
          <p className="empty">{activeQuery ? "По вашему запросу ничего не найдено." : "Продуктов пока нет. Добавьте первый."}</p>
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
