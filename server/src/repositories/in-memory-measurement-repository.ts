import { randomUUID } from "node:crypto";
import type { Measurement, NewMeasurement } from "../domain/measurement.js";
import type { MeasurementRepository } from "./measurement-repository.js";

const seed: Measurement[] = [
  { id: "sample-weight", type: "weight", value: 72.4, unit: "кг", measuredAt: "2026-09-12T08:00:00.000Z" },
  { id: "sample-pulse", type: "pulse", value: 68, unit: "уд/мин", measuredAt: "2026-09-13T07:30:00.000Z" }
];

export class InMemoryMeasurementRepository implements MeasurementRepository {
  private readonly measurements: Measurement[];

  constructor(initial: Measurement[] = seed) {
    this.measurements = structuredClone(initial);
  }

  async findAll(): Promise<Measurement[]> {
    return structuredClone([...this.measurements].sort((a, b) => b.measuredAt.localeCompare(a.measuredAt)));
  }

  async create(data: NewMeasurement): Promise<Measurement> {
    const measurement = { id: randomUUID(), ...data };
    this.measurements.push(measurement);
    return structuredClone(measurement);
  }
}
