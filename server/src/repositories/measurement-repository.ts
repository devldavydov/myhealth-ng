import type { Measurement, NewMeasurement } from "../domain/measurement.js";

// Маршруты зависят от контракта, а не от способа хранения. Реализацию можно
// заменить на PostgreSQL-репозиторий без изменений в API.
export interface MeasurementRepository {
  findAll(): Promise<Measurement[]>;
  create(measurement: NewMeasurement): Promise<Measurement>;
}
