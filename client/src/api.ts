export type MeasurementType = "weight" | "pressure" | "pulse";
export interface Measurement {
  id: string;
  type: MeasurementType;
  value: number;
  unit: string;
  measuredAt: string;
}
export type NewMeasurement = Omit<Measurement, "id">;
export interface CurrentUser { guid: string; name: string; }

async function apiRequest<T>(url: string, init?: RequestInit): Promise<T> {
  const response = await fetch(url, init);
  const body = (await response.json()) as T & { error?: string };
  if (!response.ok) throw new Error(body.error ?? "Не удалось выполнить запрос");
  return body;
}

export async function getCurrentUser(): Promise<CurrentUser> {
  return (await apiRequest<{ data: CurrentUser }>("/api/me")).data;
}

export async function getMeasurements(): Promise<Measurement[]> {
  return (await apiRequest<{ data: Measurement[] }>("/api/measurements")).data;
}

export async function createMeasurement(data: NewMeasurement): Promise<Measurement> {
  return (await apiRequest<{ data: Measurement }>("/api/measurements", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(data)
  })).data;
}
