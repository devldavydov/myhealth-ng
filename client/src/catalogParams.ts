import type { PageSize } from "./api";

export const DEFAULT_PAGE = 1;
export const DEFAULT_PAGE_SIZE: PageSize = 20;

export type CatalogParams = {
  query: string;
  page: number;
  pageSize: PageSize;
};

export function readCatalogParams(searchParams: URLSearchParams): CatalogParams {
  const pageValue = Number(searchParams.get("page"));
  const pageSizeValue = Number(searchParams.get("pageSize"));
  return {
    query: searchParams.get("q")?.trim() ?? "",
    page: Number.isInteger(pageValue) && pageValue >= 1 ? pageValue : DEFAULT_PAGE,
    pageSize: pageSizeValue === 10 || pageSizeValue === 20 || pageSizeValue === 50 ? pageSizeValue : DEFAULT_PAGE_SIZE
  };
}

export function catalogSearchParams(params: CatalogParams): URLSearchParams {
  const result = new URLSearchParams();
  if (params.query) result.set("q", params.query);
  if (params.page !== DEFAULT_PAGE) result.set("page", String(params.page));
  if (params.pageSize !== DEFAULT_PAGE_SIZE) result.set("pageSize", String(params.pageSize));
  return result;
}
