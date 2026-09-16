export interface Food {
  key: string;
  name: string;
  brand: string;
  cal100: number;
  prot100: number;
  fat100: number;
  carb100: number;
  comment: string;
}

export type FoodData = Omit<Food, "key">;

export type PageSize = 10 | 20 | 50;

export interface Pagination {
  page: number;
  pageSize: PageSize;
  total: number;
  totalPages: number;
}

export interface Page<T> {
  items: T[];
  pagination: Pagination;
}

export interface PageQuery {
  query?: string;
  page?: number;
  pageSize?: PageSize;
}

export interface BundleTotals {
  weight: number;
  cal: number;
  protein: number;
  fat: number;
  carb: number;
}

export interface BundleSummary {
  key: string;
  name: string;
  itemCount: number;
  totals: BundleTotals;
}

export interface BundleItem {
  food: Food;
  weight: number;
}

export interface Bundle {
  key: string;
  name: string;
  items: BundleItem[];
  totals: BundleTotals;
}

export interface BundleData {
  name: string;
  items: Array<{ foodKey: string; weight: number }>;
}
export interface WeightEntry {
  dt: string;
  value: number;
}

export type WeightData = WeightEntry;
export interface CurrentUser { guid: string; name: string; }

export interface UserSettings {
  defaultDailyCalorieLimit: number | null;
}

export interface UserSettingsData {
  defaultDailyCalorieLimit: number;
}

export interface ActiveCalories {
  dt: string;
  value: number;
}

export type ActiveCaloriesData = ActiveCalories;

export type MealType = "завтрак" | "до обеда" | "обед" | "полдник" | "до ужина" | "ужин";

export interface JournalTotals {
  weight: number;
  cal: number;
  protein: number;
  fat: number;
  carb: number;
}

export interface MacroPercent {
  protein: number;
  fat: number;
  carb: number;
}

export interface JournalItem {
  food: Food;
  weight: number;
}

export interface JournalZone {
  meal: MealType;
  items: JournalItem[];
  totals: JournalTotals;
}

export interface JournalDay {
  dt: string;
  zones: JournalZone[];
  totals: JournalTotals;
  macroPercent: MacroPercent;
}

export interface JournalSaveData {
  dt: string;
  meal: MealType;
  items: Array<{ foodKey: string; weight: number }>;
}

export class ApiError extends Error {
  constructor(message: string, readonly details: Record<string, string[]> = {}) {
    super(message);
    this.name = "ApiError";
  }
}

async function apiRequest<T>(url: string, init?: RequestInit): Promise<T> {
  const response = await fetch(url, init);
  const body = (await response.json()) as T & { error?: string; details?: Record<string, string[]> };
  if (!response.ok) {
    const details = body.details ? Object.values(body.details).flat().join(". ") : "";
    throw new ApiError([body.error ?? "Не удалось выполнить запрос", details].filter(Boolean).join(": "), body.details);
  }
  return body;
}

async function getPage<T>(url: string, query: PageQuery): Promise<Page<T>> {
  const params = new URLSearchParams();
  const normalizedQuery = query.query?.trim();
  if (normalizedQuery) params.set("q", normalizedQuery);
  params.set("page", String(query.page ?? 1));
  params.set("pageSize", String(query.pageSize ?? 20));

  const body = await apiRequest<{ data: T[]; pagination: Pagination }>(url + "?" + params.toString());
  return { items: body.data, pagination: body.pagination };
}

export async function getCurrentUser(): Promise<CurrentUser> {
  return (await apiRequest<{ data: CurrentUser }>("/api/me")).data;
}

export async function getSettings(): Promise<UserSettings> {
  return (await apiRequest<{ data: UserSettings }>("/api/settings")).data;
}

export async function saveSettings(data: UserSettingsData): Promise<UserSettings> {
  return (await apiRequest<{ data: UserSettings }>("/api/settings", {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(data)
  })).data;
}

export async function getActiveCalories(dt: string): Promise<ActiveCalories | null> {
  return (await apiRequest<{ data: ActiveCalories | null }>(`/api/active-calories?dt=${encodeURIComponent(dt)}`)).data;
}

export async function saveActiveCalories(data: ActiveCaloriesData): Promise<ActiveCalories> {
  return (await apiRequest<{ data: ActiveCalories }>("/api/active-calories", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(data)
  })).data;
}

export async function deleteActiveCalories(dt: string): Promise<void> {
  const response = await fetch(`/api/active-calories/${encodeURIComponent(dt)}`, { method: "DELETE" });
  if (response.ok) return;
  let message = "Не удалось удалить активные калории";
  try {
    const body = (await response.json()) as { error?: string };
    message = body.error ?? message;
  } catch {
    // Keep the fallback for a non-JSON proxy error.
  }
  throw new Error(message);
}

export async function getJournal(dt: string): Promise<JournalDay> {
  return (await apiRequest<{ data: JournalDay }>(`/api/journal?dt=${encodeURIComponent(dt)}`)).data;
}

export async function saveJournal(data: JournalSaveData): Promise<JournalDay> {
  return (await apiRequest<{ data: JournalDay }>("/api/journal", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(data)
  })).data;
}

export async function deleteJournalItem(dt: string, meal: MealType, foodKey: string): Promise<JournalDay> {
  return (await apiRequest<{ data: JournalDay }>(`/api/journal/${encodeURIComponent(dt)}/${encodeURIComponent(meal)}/${encodeURIComponent(foodKey)}`, {
    method: "DELETE"
  })).data;
}

export async function clearJournalMeal(dt: string, meal: MealType): Promise<JournalDay> {
  return (await apiRequest<{ data: JournalDay }>(`/api/journal/${encodeURIComponent(dt)}/${encodeURIComponent(meal)}`, {
    method: "DELETE"
  })).data;
}

export async function getFood(query: PageQuery = {}): Promise<Page<Food>> {
  return getPage<Food>("/api/food", query);
}

export async function getFoodByKey(key: string): Promise<Food> {
  return (await apiRequest<{ data: Food }>(`/api/food/${encodeURIComponent(key)}`)).data;
}

export async function createFood(data: FoodData): Promise<Food> {
  return (await apiRequest<{ data: Food }>("/api/food", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(data)
  })).data;
}

export async function updateFood(key: string, data: FoodData): Promise<Food> {
  return (await apiRequest<{ data: Food }>(`/api/food/${encodeURIComponent(key)}`, {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(data)
  })).data;
}

export async function deleteFood(key: string): Promise<void> {
  const response = await fetch(`/api/food/${encodeURIComponent(key)}`, { method: "DELETE" });
  if (response.ok) return;
  let message = "Не удалось удалить продукт";
  try {
    const body = (await response.json()) as { error?: string };
    message = body.error ?? message;
  } catch {
    // Keep the fallback for a non-JSON proxy error.
  }
  throw new Error(message);
}

export async function getBundles(query: PageQuery = {}): Promise<Page<BundleSummary>> {
  return getPage<BundleSummary>("/api/bundle", query);
}

export async function getBundleByKey(key: string): Promise<Bundle> {
  return (await apiRequest<{ data: Bundle }>(`/api/bundle/${encodeURIComponent(key)}`)).data;
}

export async function createBundle(data: BundleData): Promise<Bundle> {
  return (await apiRequest<{ data: Bundle }>("/api/bundle", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(data)
  })).data;
}

export async function updateBundle(key: string, data: BundleData): Promise<Bundle> {
  return (await apiRequest<{ data: Bundle }>(`/api/bundle/${encodeURIComponent(key)}`, {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(data)
  })).data;
}

export async function deleteBundle(key: string): Promise<void> {
  const response = await fetch(`/api/bundle/${encodeURIComponent(key)}`, { method: "DELETE" });
  if (response.ok) return;
  let message = "Не удалось удалить бандл";
  try {
    const body = (await response.json()) as { error?: string };
    message = body.error ?? message;
  } catch {
    // Keep the fallback for a non-JSON proxy error.
  }
  throw new Error(message);
}

export async function getWeight(from = "", to = ""): Promise<WeightEntry[]> {
  const params = new URLSearchParams();
  if (from) params.set("from", from);
  if (to) params.set("to", to);
  const suffix = params.size > 0 ? `?${params.toString()}` : "";
  return (await apiRequest<{ data: WeightEntry[] }>(`/api/weight${suffix}`)).data;
}

export async function saveWeight(data: WeightData): Promise<WeightEntry> {
  return (await apiRequest<{ data: WeightEntry }>("/api/weight", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(data)
  })).data;
}

export async function deleteWeight(dt: string): Promise<void> {
  const response = await fetch(`/api/weight/${encodeURIComponent(dt)}`, { method: "DELETE" });
  if (response.ok) return;
  let message = "Не удалось удалить запись веса";
  try {
    const body = (await response.json()) as { error?: string };
    message = body.error ?? message;
  } catch {
    // Keep the fallback for a non-JSON proxy error.
  }
  throw new Error(message);
}
