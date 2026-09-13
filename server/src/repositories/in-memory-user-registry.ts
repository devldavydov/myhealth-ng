import type { UserIdentity } from "../domain/user.js";
import type { UserRegistry } from "./user-registry.js";

export class InMemoryUserRegistry implements UserRegistry {
  readonly users = new Map<string, UserIdentity>();

  async remember(user: UserIdentity): Promise<void> {
    this.users.set(user.guid, structuredClone(user));
  }
}
