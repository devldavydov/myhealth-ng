import { X509Certificate } from "node:crypto";
import type { Request } from "express";
import type { UserIdentity } from "../domain/user.js";

const guidPattern = "[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}";

export class ClientCertificateError extends Error {}

function header(request: Request, name: string): string | undefined {
  const value = request.headers[name];
  return Array.isArray(value) ? value[0] : value;
}

export function readClientIdentity(request: Request, certificateRequired: boolean): UserIdentity {
  const verification = header(request, "x-client-verify");
  const escapedCertificate = header(request, "x-client-cert");

  if (verification !== "SUCCESS" || !escapedCertificate) {
    if (certificateRequired) throw new ClientCertificateError("Требуется проверенный клиентский сертификат");
    return {
      guid: "00000000-0000-4000-8000-000000000000",
      name: "Локальный пользователь",
      certificateFingerprint: "development",
      lastSeenAt: new Date().toISOString()
    };
  }

  try {
    const certificate = new X509Certificate(decodeURIComponent(escapedCertificate));
    const guid = certificate.subjectAltName?.match(new RegExp(`URI:urn:myhealth:user:(${guidPattern})`, "i"))?.[1];
    const name = certificate.subject.match(/(?:^|\n)CN=([^\n]+)(?:\n|$)/)?.[1];
    if (!guid || !name) throw new Error("В сертификате отсутствуют CN или MyHealth GUID");

    return {
      guid: guid.toLowerCase(),
      name,
      certificateFingerprint: certificate.fingerprint256,
      lastSeenAt: new Date().toISOString()
    };
  } catch (error) {
    throw new ClientCertificateError(
      error instanceof Error ? `Некорректный клиентский сертификат: ${error.message}` : "Некорректный клиентский сертификат"
    );
  }
}
