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
export interface WeightEntry {
  dt: string;
  value: number;
}

export type WeightData = WeightEntry;
export interface CurrentUser { guid: string; name: string; }

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

export async function getCurrentUser(): Promise<CurrentUser> {
  return (await apiRequest<{ data: CurrentUser }>("/api/me")).data;
}

export async function getFood(query = ""): Promise<Food[]> {
  const suffix = query.trim() ? `?q=${encodeURIComponent(query.trim())}` : "";
  return (await apiRequest<{ data: Food[] }>(`/api/food${suffix}`)).data;
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
