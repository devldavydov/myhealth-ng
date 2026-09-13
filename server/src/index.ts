import { createApp } from "./app.js";
import { InMemoryMeasurementRepository } from "./repositories/in-memory-measurement-repository.js";

const port = Number(process.env.PORT ?? 3000);
const host = process.env.HOST ?? "127.0.0.1";

createApp(new InMemoryMeasurementRepository()).listen(port, host, () => {
  console.log(`MyHealth API: http://${host}:${port}`);
});
