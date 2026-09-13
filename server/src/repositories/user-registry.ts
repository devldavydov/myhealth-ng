import type { UserIdentity } from "../domain/user.js";

export interface UserRegistry {
  remember(user: UserIdentity): Promise<void>;
}
