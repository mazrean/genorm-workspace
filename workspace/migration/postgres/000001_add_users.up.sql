CREATE TABLE "users" (
  "id"  CHAR(36) PRIMARY KEY NOT NULL,
  "name" VARCHAR(64) NOT NULL,
  "created_at" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
