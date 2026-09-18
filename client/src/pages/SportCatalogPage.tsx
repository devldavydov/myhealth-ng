import { type FormEvent, useCallback, useEffect, useMemo, useRef, useState } from "react";
import { Link, useSearchParams } from "react-router-dom";
import { getSports, type Pagination, type Sport } from "../api";
import { catalogSearchParams, readCatalogParams, type CatalogParams } from "../catalogParams";
import { CatalogPagination } from "../components/CatalogPagination";
import { ConfirmDialog } from "../components/ConfirmDialog";

const initialPagination: Pagination = { page: 1, pageSize: 20, total: 0, totalPages: 0 };

export function SportCatalogPage() {
  const [searchParams, setSearchParams] = useSearchParams();
  const catalog = useMemo(() => readCatalogParams(searchParams), [searchParams]);
  const [query, setQuery] = useState(catalog.query);
  const [items, setItems] = useState<Sport[]>([]);
  const [pagination, setPagination] = useState(initialPagination);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [target, setTarget] = useState<Sport | null>(null);
  const [deleting, setDeleting] = useState(false);
  const sequence = useRef(0);
  const navigate = useCallback((params: CatalogParams) => setSearchParams(catalogSearchParams(params)), [setSearchParams]);
  const load = useCallback(async (params: CatalogParams) => {
    const request = ++sequence.current; setLoading(true); setError("");
    try { const page = await getSports(params); if (request === sequence.current) { setItems(page.items); setPagination(page.pagination); } }
    catch (requestError) { if (request === sequence.current) setError(requestError instanceof Error ? requestError.message : "Не удалось загрузить виды спорта"); }
    finally { if (request === sequence.current) setLoading(false); }
  }, []);
  useEffect(() => { setQuery(catalog.query); void load(catalog); return () => { sequence.current += 1; }; }, [catalog, load]);
  function search(event: FormEvent) { event.preventDefault(); navigate({ ...catalog, query: query.trim(), page: 1 }); }
  async function remove() { if (!target) return; setDeleting(true); try { const response = await fetch(`/api/sport/${encodeURIComponent(target.key)}`, { method: "DELETE" }); if (!response.ok) { const body = await response.json() as { error?: string }; throw new Error(body.error ?? "Не удалось удалить вид спорта"); } setTarget(null); await load(catalog); } catch (requestError) { setError(requestError instanceof Error ? requestError.message : "Не удалось удалить вид спорта"); } finally { setDeleting(false); } }
  return <div className="page"><header className="page-header"><div><p className="eyebrow">Справочник</p><h1>Спорт</h1><p className="page-description">Общий справочник видов активности и единиц измерения.</p></div><Link className="button primary" to="/sport/new">Добавить вид</Link></header><section className="food-catalog"><div className="section-heading"><h2>Список</h2><span className="food-count" aria-label={`Найдено видов спорта: ${pagination.total}`}>{pagination.total}</span></div><form className="search-form" role="search" onSubmit={search}><input aria-label="Поиск видов спорта" value={query} onChange={(event) => setQuery(event.target.value)} placeholder="Название"/><button className="button primary">Найти</button></form>{error && <div className="error" role="alert">{error}</div>}{loading ? <p className="muted">Загрузка…</p> : items.length === 0 ? <p className="empty">Виды спорта не найдены.</p> : <div className="food-list">{items.map((sport) => <article className="food-card" key={sport.key}><div className="food-card-heading"><div><h3>{sport.name}</h3><p>{sport.comment || "Без комментария"}</p></div><strong><small>{sport.unit}</small></strong></div><div className="food-actions"><Link className="text-button" to={`/sport/${encodeURIComponent(sport.key)}/edit`}>Редактировать</Link><button className="text-button danger" onClick={() => setTarget(sport)} type="button">Удалить</button></div></article>)}</div>}<CatalogPagination noun="видов спорта" pagination={pagination} onPageChange={(page) => navigate({ ...catalog, page })} onPageSizeChange={(pageSize) => navigate({ ...catalog, page: 1, pageSize })}/></section><ConfirmDialog open={target !== null} title="Удалить вид спорта?" confirmLabel="Удалить" busy={deleting} dangerous onCancel={() => setTarget(null)} onConfirm={() => void remove()}>Вид «{target?.name}» будет удалён. Если он используется в активности, сервер отклонит удаление.</ConfirmDialog></div>;
}
