export const measurementTypes = ["weight", "pressure", "pulse"] as const;
export type MeasurementType = (typeof measurementTypes)[number];

export interface Measurement {
  id: string;
  type: MeasurementType;
  value: number;
  unit: string;
  measuredAt: string;
}

export type NewMeasurement = Omit<Measurement, "id">;
