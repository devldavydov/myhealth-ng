import { Router } from "express";
import { z } from "zod";
import { measurementTypes } from "../domain/measurement.js";
import type { MeasurementRepository } from "../repositories/measurement-repository.js";

const schema = z.object({
  type: z.enum(measurementTypes),
  value: z.number().finite(),
  unit: z.string().trim().min(1).max(20),
  measuredAt: z.iso.datetime()
});

export function createMeasurementsRouter(repository: MeasurementRepository): Router {
  const router = Router();

  router.get("/", async (_request, response, next) => {
    try {
      response.json({ data: await repository.findAll() });
    } catch (error) {
      next(error);
    }
  });

  router.post("/", async (request, response, next) => {
    const parsed = schema.safeParse(request.body);
    if (!parsed.success) {
      response.status(400).json({
        error: "Некорректные данные измерения",
        details: z.flattenError(parsed.error).fieldErrors
      });
      return;
    }
    try {
      response.status(201).json({ data: await repository.create(parsed.data) });
    } catch (error) {
      next(error);
    }
  });

  return router;
}
