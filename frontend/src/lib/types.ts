export type UserPublic = {
  id: number;
  email: string;
};

export type TokenPayload = {
  accessToken: string;
  expiresAt: string;
  user: UserPublic;
};

export type MePayload = {
  id: number;
  email: string;
  message: string;
};

export type TransferRecord = {
  id: number;
  fromUserId: number;
  toUserId: number;
  senderEmail: string;
  recipientEmail: string;
  amount: string;
  notes: string;
  direction: "in" | "out";
  createdAt: string;
};

export type WalletPayload = {
  balance: string;
  transfers: TransferRecord[];
};

export type ApiErrorBody = {
  error: {
    code: string;
    message: string;
    details?: Record<string, string>;
  };
};

export type QueuedTransfer = {
  id: string;
  userId: number;
  recipient: string;
  amount: string;
  notes: string;
  status: "pending" | "synced" | "failed";
  error?: string;
  createdAt: number;
  attempts: number;
};
