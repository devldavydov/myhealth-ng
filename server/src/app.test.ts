import request from "supertest";
import { describe, expect, it } from "vitest";
import { createApp } from "./app.js";
import { InMemoryMeasurementRepository } from "./repositories/in-memory-measurement-repository.js";

const app = () => createApp(new InMemoryMeasurementRepository([]));

describe("MyHealth API", () => {
  it("возвращает состояние сервиса", async () => {
    const response = await request(app()).get("/api/health");
    expect(response.status).toBe(200);
    expect(response.body).toEqual({ status: "ok" });
  });

  it("создаёт и возвращает измерение", async () => {
    const measurement = { type: "weight", value: 71.8, unit: "кг", measuredAt: "2026-09-13T10:00:00.000Z" };
    const response = await request(app()).post("/api/measurements").send(measurement);
    expect(response.status).toBe(201);
    expect(response.body.data).toMatchObject(measurement);
    expect(response.body.data.id).toEqual(expect.any(String));
  });

  it("отклоняет некорректное измерение", async () => {
    const response = await request(app()).post("/api/measurements").send({ type: "unknown", value: "много" });
    expect(response.status).toBe(400);
    expect(response.body.error).toBe("Некорректные данные измерения");
  });
});
