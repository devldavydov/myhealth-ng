import type { PageSize, Pagination } from "../api";

type CatalogPaginationProps = {
  noun: string;
  onPageChange: (page: number) => void;
  onPageSizeChange: (pageSize: PageSize) => void;
  pagination: Pagination;
};

export function CatalogPagination({ noun, onPageChange, onPageSizeChange, pagination }: CatalogPaginationProps) {
  const hasPages = pagination.totalPages > 0;
  return (
    <nav aria-label={`Пагинация ${noun}`} className="catalog-pagination">
      <label>На странице
        <select
          aria-label={`Количество ${noun} на странице`}
          value={pagination.pageSize}
          onChange={(event) => onPageSizeChange(Number(event.target.value) as PageSize)}
        >
          <option value={10}>10</option>
          <option value={20}>20</option>
          <option value={50}>50</option>
        </select>
      </label>
      <div className="catalog-pagination-nav">
        <button
          aria-label="Предыдущая страница"
          className="button secondary"
          disabled={!hasPages || pagination.page <= 1}
          onClick={() => onPageChange(pagination.page - 1)}
          type="button"
        >
          Назад
        </button>
        <span>{hasPages ? `Страница ${pagination.page} из ${pagination.totalPages}` : "Нет страниц"}</span>
        <button
          aria-label="Следующая страница"
          className="button secondary"
          disabled={!hasPages || pagination.page >= pagination.totalPages}
          onClick={() => onPageChange(pagination.page + 1)}
          type="button"
        >
          Вперёд
        </button>
      </div>
    </nav>
  );
}
