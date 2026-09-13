import cors from "cors";
import express, { type ErrorRequestHandler } from "express";
import { ClientCertificateError, readClientIdentity } from "./auth/client-identity.js";
import { InMemoryUserRegistry } from "./repositories/in-memory-user-registry.js";
import type { MeasurementRepository } from "./repositories/measurement-repository.js";
import type { UserRegistry } from "./repositories/user-registry.js";
import { createMeasurementsRouter } from "./routes/measurements.js";

export function createApp(
  repository: MeasurementRepository,
  userRegistry: UserRegistry = new InMemoryUserRegistry()
) {
  const app = express();
  const certificateRequired = process.env.REQUIRE_CLIENT_CERT === "true";

  app.disable("x-powered-by");
  app.use(cors({ origin: process.env.CLIENT_ORIGIN ?? "http://localhost:5173" }));
  app.use(express.json({ limit: "100kb" }));
  app.use(async (request, response, next) => {
    try {
      const identity = readClientIdentity(request, certificateRequired);
      await userRegistry.remember(identity);
      response.locals.user = identity;
      next();
    } catch (error) {
      next(error);
    }
  });

  app.get("/api/health", (_request, response) => response.json({ status: "ok" }));
  app.get("/api/me", (_request, response) => response.json({ data: response.locals.user }));
  app.use("/api/measurements", createMeasurementsRouter(repository));
  app.use((_request, response) => response.status(404).json({ error: "Маршрут не найден" }));

  const errorHandler: ErrorRequestHandler = (error, _request, response, _next) => {
    if (error instanceof ClientCertificateError) {
      response.status(403).json({ error: error.message });
      return;
    }
    console.error(error);
    response.status(500).json({ error: "Внутренняя ошибка сервера" });
  };
  app.use(errorHandler);
  return app;
}
