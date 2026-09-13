import { execFileSync } from "node:child_process";
import { mkdtempSync, readFileSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";
import type { Request } from "express";
import { describe, expect, it } from "vitest";
import { readClientIdentity } from "./client-identity.js";

describe("client certificate identity", () => {
  it("разбирает имя и GUID из сертификата", () => {
    const directory = mkdtempSync(join(tmpdir(), "myhealth-cert-"));
    const key = join(directory, "key.pem");
    const certificate = join(directory, "cert.pem");
    const guid = "3f67c05f-7c9e-4cb5-b26a-f9ce5b065865";
    execFileSync("openssl", ["req", "-utf8", "-x509", "-newkey", "rsa:2048", "-nodes", "-days", "1", "-keyout", key, "-out", certificate, "-subj", "/CN=Иван Иванов", "-addext", `subjectAltName=URI:urn:myhealth:user:${guid}`], { stdio: "ignore" });
    const pem = readFileSync(certificate, "utf8");
    const request = { headers: { "x-client-verify": "SUCCESS", "x-client-cert": encodeURIComponent(pem) } } as unknown as Request;

    expect(readClientIdentity(request, true)).toMatchObject({ guid, name: "Иван Иванов" });
  });

  it("не принимает запрос без сертификата в production-режиме", () => {
    expect(() => readClientIdentity({ headers: {} } as Request, true)).toThrow("проверенный клиентский сертификат");
  });
});
